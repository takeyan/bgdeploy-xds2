package bgdeploy

import (
        "context"

//      "reflect"
        "strings"
        "fmt"
        "net/http"
//        "k8s.io/apimachinery/pkg/labels"

        appsv1 "k8s.io/api/apps/v1"            // add for Depoyment
        "k8s.io/apimachinery/pkg/util/intstr"  // add for TargetPort
        swallowlabv1alpha1 "bgdeploy-xds2/pkg/apis/swallowlab/v1alpha1"      // bgdeploy-xds2

        corev1 "k8s.io/api/core/v1"
        "k8s.io/apimachinery/pkg/api/errors"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/apimachinery/pkg/runtime"
        "k8s.io/apimachinery/pkg/types"
        "sigs.k8s.io/controller-runtime/pkg/client"
        "sigs.k8s.io/controller-runtime/pkg/controller"
        "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
        "sigs.k8s.io/controller-runtime/pkg/handler"
        logf "sigs.k8s.io/controller-runtime/pkg/log"
        "sigs.k8s.io/controller-runtime/pkg/manager"
        "sigs.k8s.io/controller-runtime/pkg/reconcile"
        "sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_bgdeploy")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new BGDeploy Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
        return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
        return &ReconcileBGDeploy{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
        // Create a new controller
        c, err := controller.New("bgdeploy-controller", mgr, controller.Options{Reconciler: r})
        if err != nil {
                return err
        }

        // Watch for changes to primary resource BGDeploy
        err = c.Watch(&source.Kind{Type: &swallowlabv1alpha1.BGDeploy{}}, &handler.EnqueueRequestForObject{})
        if err != nil {
                return err
        }


// WATCH BLOCK BEGIN
        // TODO(user): Modify this to be the types you create that are owned by the primary resource
        // Watch for changes to secondary resource Pods and requeue the owner EchoFlask


        err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
                IsController: true,
                OwnerType: &swallowlabv1alpha1.BGDeploy{},
        })
        if err != nil {
                return err
        }

        err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
                IsController: true,
                OwnerType: &swallowlabv1alpha1.BGDeploy{},
        })
        if err != nil {
                return err
        }

// WATCH BLOCK END
        return nil
}

// blank assignment to verify that ReconcileBGDeploy implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileBGDeploy{}

// ReconcileBGDeploy reconciles a BGDeploy object
type ReconcileBGDeploy struct {
        // This client, initialized using mgr.Client() above, is a split client
        // that reads objects from the cache and writes to the apiserver
        client client.Client
        scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a BGDeploy object and makes changes based on the state read
// and what is in the BGDeploy.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBGDeploy) Reconcile(request reconcile.Request) (reconcile.Result, error) {
        reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
        reqLogger.Info("Reconciling BGDeploy")

        // Fetch the BGDeploy instance
        instance := &swallowlabv1alpha1.BGDeploy{}
        err := r.client.Get(context.TODO(), request.NamespacedName, instance)
        if err != nil {
                if errors.IsNotFound(err) {
                        // Request object not found, could have been deleted after reconcile request.
                        // Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
                        // Return and don't requeue
                        return reconcile.Result{}, nil
                }
                // Error reading the object - requeue the request.
                return reconcile.Result{}, err
        }

// RECONCILE BLOCK BEGIN

// Reconcilation Policy:
//
// Always
//      bring-up  bgdeploy-pod-envoy, bgdeploy-svc-xds, and bgdeploy-svc-envoy if they are not running there

// IF active==blue && transit==OFF {
//      bring-up  bgdeploy-dep-blue  and bgdeploy-svc-blue  if they are not running there
//      terminame bgdeploy-dep-green and bgdeploy-svc-green if they are running there
//      }

// ELSE IF active==blue && transit==ON  {
//      bring-up  bgdeploy-dep-blue  and bgdeploy-svc-blue  if they are not running there
//      bring-up  bgdeploy-dep-green and bgdeploy-svc-green if they are not running there
//      update XDS snapshot with direction=blue  if the direction is not blue
//      }

// ELSE IF active==green && transit==ON  {
//      bring-up  bgdeploy-dep-blue  and bgdeploy-svc-blue  if they are not running there
//      bring-up  bgdeploy-dep-green and bgdeploy-svc-green if they are not running there
//      update XDS snapshot with direction=green if the direction is not green
//      }

// ELSE IF active==green && transit==OFF {
//      terminate bgdeploy-dep-blue  and bgdeploy-svc-blue  if they are not running there
//      bring-up  bgdeploy-dep-green and bgdeploy-svc-green if they are not running there
//      }


        blueImage   := instance.Spec.Blue
        greenImage  := instance.Spec.Green
//              port        := instance.Spec.Port
//              replicas    := instance.Spec.Replicas
        transitFlag := strings.ToUpper(instance.Spec.Transit)   // ON or OFF
        activeApp   := strings.ToUpper(instance.Spec.Active)      // BLUE or GREEN

//                ctx := context.TODO()

//                pods := &corev1.PodList{}
//              deps := &appsv1.DeploymentList{}
//                svcs := &corev1.ServiceList{}



// Instance Name
        bgdeploy_pod_envoy  := instance.Name + "-pod-envoy"
//        bgdeploy_pod_blue   := instance.Name + "-pod-blue"
//              bgdeploy_pod_green  := instance.Name + "-pod-green"
        bgdeploy_dep_blue   := instance.Name + "-dep-blue"
        bgdeploy_dep_green  := instance.Name + "-dep-green"
        bgdeploy_svc_xds    := instance.Name + "-svc-xds"
        bgdeploy_svc_envoy  := instance.Name + "-svc-envoy"
        bgdeploy_svc_blue   := instance.Name + "-svc-blue"
        bgdeploy_svc_green  := instance.Name + "-svc-green"


// Label
        l_bgdeploy_pod_envoy  := map[string]string{ "app" : "bgdeploy" , "service" : "envoy" }
//        l_bgdeploy_pod_blue   := map[string]string{ "app" : "bgdeploy" , "color"   : "blue"  }
//              l_bgdeploy_pod_green  := map[string]string{ "app" : "bgdeploy" , "color"   : "green" }
        l_bgdeploy_dep_blue   := map[string]string{ "app" : "bgdeploy" , "color"   : "blue"  }
        l_bgdeploy_dep_green  := map[string]string{ "app" : "bgdeploy" , "color"   : "green" }
        l_bgdeploy_svc_xds    := map[string]string{ "app" : "bgdeploy" , "service" : "xds"   }
        l_bgdeploy_svc_envoy  := map[string]string{ "app" : "bgdeploy" , "service" : "envoy" }
        l_bgdeploy_svc_blue   := map[string]string{ "app" : "bgdeploy" , "color"   : "blue"  }
        l_bgdeploy_svc_green  := map[string]string{ "app" : "bgdeploy" , "color"   : "green" }


// --------------------------
        podfound := &corev1.Pod{}
        depfound := &appsv1.Deployment{}
        svcfound := &corev1.Service{}
        xdsfound := &corev1.Service{}


// Always
//      bring-up  bgdeploy-pod-envoy, bgdeploy-svc-xds, and bgdeploy-svc-envoy if they are not running there

        // Check if the envoy pod already exists, if not create a new one
        envoyPod := r.newEnvoyPodForCR(instance, bgdeploy_pod_envoy, l_bgdeploy_pod_envoy)
//        podfound := &corev1.Pod{}
        err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_pod_envoy, Namespace: instance.Namespace}, podfound)
        if err != nil && errors.IsNotFound(err) {
            reqLogger.Info("Creating envoy Pod", "Pod.Namespace", envoyPod.Namespace, "Pod.Name", envoyPod.Name)
            err = r.client.Create(context.TODO(), envoyPod)
        }



        // Check if the envoy service already exists, if not create a new one
        envoySvc := r.newEnvoyServiceForCR(instance, bgdeploy_svc_envoy, l_bgdeploy_svc_envoy)
//        svcfound := &corev1.Service{}
        err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_svc_envoy, Namespace: instance.Namespace}, svcfound)
        if err != nil && errors.IsNotFound(err) {
            reqLogger.Info("Creating envoy Service", "Svc.Namespace", envoySvc.Namespace, "Svc.Name", envoySvc.Name)
            err = r.client.Create(context.TODO(), envoySvc)
        }




        // Check if the xds service already exists, if not create a new one
        xdsSvc := r.newXDSServiceForCR(instance, bgdeploy_svc_xds, l_bgdeploy_svc_xds)
//        xdsfound := &corev1.Service{}
        err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_svc_xds, Namespace: instance.Namespace}, xdsfound)
        if err != nil && errors.IsNotFound(err) {
            reqLogger.Info("Creating xds Service", "Xds.Namespace", xdsSvc.Namespace, "Xds.Name", xdsSvc.Name)
            err = r.client.Create(context.TODO(), xdsSvc)
        }




// IF active==blue && transit==OFF {
//      bring-up  bgdeploy-dep-blue  and bgdeploy-svc-blue  if they are not running there
//      terminame bgdeploy-dep-green and bgdeploy-svc-green if they are running there
//      }
    if activeApp == "BLUE" && transitFlag == "OFF" {

        // Check if the blue deployment already exists, if not create a new one
        err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_dep_blue, Namespace: instance.Namespace}, depfound)
        if err != nil && errors.IsNotFound(err) {
            // Define a new deployment
            reqLogger.Info("Defining a new Deployment for: " + bgdeploy_dep_blue)
            blueDep := r.newBGDeploymentForCR(instance, bgdeploy_dep_blue, blueImage, l_bgdeploy_dep_blue)
            reqLogger.Info("Creating a App Deployment", "Deployment.Namespace", blueDep.Namespace, "Deployment.Name", blueDep.Name)
            err = r.client.Create(context.TODO(), blueDep)
            if err != nil {
                reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", blueDep.Namespace, "Deployment.Name", blueDep.Name)
                return reconcile.Result{}, err
            }
            // Deployment created successfully - return and requeue
            return reconcile.Result{Requeue: true}, nil
        } else if err != nil {
            reqLogger.Error(err, "Failed to get Deployment")
            return reconcile.Result{}, err
        }


        // Check if the blue service already exists, if not create a new one
        err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_svc_blue, Namespace: instance.Namespace}, svcfound)
        if err != nil && errors.IsNotFound(err) {
            blueSvc := r.newBGServiceForCR(instance, bgdeploy_svc_blue, l_bgdeploy_svc_blue)
            reqLogger.Info("Creating xds Service", "Xds.Namespace", blueSvc.Namespace, "Xds.Name", blueSvc.Name)
            err = r.client.Create(context.TODO(), blueSvc)
            if err != nil {
                reqLogger.Error(err, "Failed to create new Service", "Deployment.Namespace", blueSvc.Namespace, "Deployment.Name", blueSvc.Name)
                return reconcile.Result{}, err
            }
            // Deployment created successfully - return and requeue
            return reconcile.Result{Requeue: true}, nil
        } else if err != nil {
            reqLogger.Error(err, "Failed to get Deployment")
            return reconcile.Result{}, err
        }


        // Check if the green deployment already exists, if yes delete that
        err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_dep_green, Namespace: instance.Namespace}, depfound)
        if err != nil && errors.IsNotFound(err) {
            reqLogger.Info( "Green Deployment not found")
        } else  {
            reqLogger.Info( "Deleting the Deployment")
            if err := r.client.Delete(context.TODO(), depfound); err != nil {
                reqLogger.Error( err, "failed to delete Deployment resource")
                return reconcile.Result{}, err
            }
            return reconcile.Result{}, err
        }


        // Check if the green service already exists, if yes delete that
        err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_svc_green, Namespace: instance.Namespace}, svcfound)
        if err != nil && errors.IsNotFound(err) {
            reqLogger.Info( "Green Service not found")
                    } else  {
            reqLogger.Info( "Deleting the Service")
            if err := r.client.Delete(context.TODO(), svcfound); err != nil {
                reqLogger.Error( err, "failed to delete Service resource")
                return reconcile.Result{}, err
            }
            return reconcile.Result{}, err
        }

        // set the upstreamURL&Port to the Blue Service
	reqURL := fmt.Sprintf("http://%s.%s:18080/xds?host=%s.%s&port=5000", bgdeploy_svc_xds, instance.Namespace, bgdeploy_svc_blue, instance.Namespace)  
        reqLogger.Info( "Changing upstreamURL: " + reqURL)
	if _, err := http.Get(reqURL); err != nil {
		reqLogger.Error( err, "failed to set the upstreamURL to Service: ", bgdeploy_svc_blue)
		return reconcile.Result{}, err
	}




    } else if activeApp == "BLUE" && transitFlag == "ON" {

// ELSE IF active==blue && transit==ON  {
//      bring-up  bgdeploy-dep-blue  and bgdeploy-svc-blue  if they are not running there
//      bring-up  bgdeploy-dep-green and bgdeploy-svc-green if they are not running there
//      update XDS snapshot with direction=blue  if the direction is not blue
//      }

        // Check if the blue deployment already exists, if not create a new one
        err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_dep_blue, Namespace: instance.Namespace}, depfound)
        if err != nil && errors.IsNotFound(err) {
            // Define a new deployment
            reqLogger.Info("Defining a new Deployment for: " + bgdeploy_dep_blue)
            blueDep := r.newBGDeploymentForCR(instance, bgdeploy_dep_blue, blueImage, l_bgdeploy_dep_blue)
            reqLogger.Info("Creating a App Deployment", "Deployment.Namespace", blueDep.Namespace, "Deployment.Name", blueDep.Name)
            err = r.client.Create(context.TODO(), blueDep)
            if err != nil {
                reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", blueDep.Namespace, "Deployment.Name", blueDep.Name)
                return reconcile.Result{}, err
            }
            // Deployment created successfully - return and requeue
            return reconcile.Result{Requeue: true}, nil
        } else if err != nil {
            reqLogger.Error(err, "Failed to get Deployment")
            return reconcile.Result{}, err
        }


        // Check if the blue service already exists, if not create a new one
        err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_svc_blue, Namespace: instance.Namespace}, svcfound)
        if err != nil && errors.IsNotFound(err) {
            blueSvc := r.newBGServiceForCR(instance, bgdeploy_svc_blue, l_bgdeploy_svc_blue)
            reqLogger.Info("Creating xds Service", "Xds.Namespace", blueSvc.Namespace, "Xds.Name", blueSvc.Name)
            err = r.client.Create(context.TODO(), blueSvc)
            if err != nil {
                reqLogger.Error(err, "Failed to create new Service", "Deployment.Namespace", blueSvc.Namespace, "Deployment.Name", blueSvc.Name)
                return reconcile.Result{}, err
            }
            // Deployment created successfully - return and requeue
            return reconcile.Result{Requeue: true}, nil
        } else if err != nil {
            reqLogger.Error(err, "Failed to get Deployment")
            return reconcile.Result{}, err
        }



        // Check if the green deployment already exists, if not create a new one
        err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_dep_green, Namespace: instance.Namespace}, depfound)
        if err != nil && errors.IsNotFound(err) {
            // Define a new deployment
            reqLogger.Info("Defining a new Deployment for: " + bgdeploy_dep_green)
            greenDep := r.newBGDeploymentForCR(instance, bgdeploy_dep_green, greenImage, l_bgdeploy_dep_green)
            reqLogger.Info("Creating a App Deployment", "Deployment.Namespace", greenDep.Namespace, "Deployment.Name", greenDep.Name)
            err = r.client.Create(context.TODO(), greenDep)
            if err != nil {
                reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", greenDep.Namespace, "Deployment.Name", greenDep.Name)
                return reconcile.Result{}, err
            }
            // Deployment created successfully - return and requeue
            return reconcile.Result{Requeue: true}, nil
        } else if err != nil {
            reqLogger.Error(err, "Failed to get Deployment")
            return reconcile.Result{}, err
        }


        // Check if the green service already exists, if not create a new one
        err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_svc_green, Namespace: instance.Namespace}, svcfound)
        if err != nil && errors.IsNotFound(err) {
            greenSvc := r.newBGServiceForCR(instance, bgdeploy_svc_green, l_bgdeploy_svc_green)
            reqLogger.Info("Creating xds Service", "Xds.Namespace", greenSvc.Namespace, "Xds.Name", greenSvc.Name)
            err = r.client.Create(context.TODO(), greenSvc)
            if err != nil {
                reqLogger.Error(err, "Failed to create new Service", "Deployment.Namespace", greenSvc.Namespace, "Deployment.Name", greenSvc.Name)
                return reconcile.Result{}, err
            }
            // Deployment created successfully - return and requeue
            return reconcile.Result{Requeue: true}, nil
        } else if err != nil {
            reqLogger.Error(err, "Failed to get Deployment")
            return reconcile.Result{}, err
        }


	// set the upstreamURL&Port to the Blue Service
	reqURL := fmt.Sprintf("http://%s.%s:18080/xds?host=%s.%s&port=5000", bgdeploy_svc_xds, instance.Namespace, bgdeploy_svc_blue, instance.Namespace)  
        reqLogger.Info( "Changing upstreamURL: " + reqURL)
	if _, err := http.Get(reqURL); err != nil {
		reqLogger.Error( err, "failed to set the upstreamURL to Service: ", bgdeploy_svc_blue)
		return reconcile.Result{}, err
	}




    } else if activeApp == "GREEN" && transitFlag == "ON" {

// ELSE IF active==green && transit==ON  {
//      bring-up  bgdeploy-dep-blue  and bgdeploy-svc-blue  if they are not running there
//      bring-up  bgdeploy-dep-green and bgdeploy-svc-green if they are not running there
//      update XDS snapshot with direction=green if the direction is not green
//      }

            // Check if the blue deployment already exists, if not create a new one
            err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_dep_blue, Namespace: instance.Namespace}, depfound)
            if err != nil && errors.IsNotFound(err) {
                // Define a new deployment
                reqLogger.Info("Defining a new Deployment for: " + bgdeploy_dep_blue)
                blueDep := r.newBGDeploymentForCR(instance, bgdeploy_dep_blue, blueImage, l_bgdeploy_dep_blue)
                reqLogger.Info("Creating a App Deployment", "Deployment.Namespace", blueDep.Namespace, "Deployment.Name", blueDep.Name)
                err = r.client.Create(context.TODO(), blueDep)
                if err != nil {
                    reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", blueDep.Namespace, "Deployment.Name", blueDep.Name)
                    return reconcile.Result{}, err
                }
                // Deployment created successfully - return and requeue
                return reconcile.Result{Requeue: true}, nil
            } else if err != nil {
                reqLogger.Error(err, "Failed to get Deployment")
                return reconcile.Result{}, err
            }


            // Check if the blue service already exists, if not create a new one
            err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_svc_blue, Namespace: instance.Namespace}, svcfound)
            if err != nil && errors.IsNotFound(err) {
                blueSvc := r.newBGServiceForCR(instance, bgdeploy_svc_blue, l_bgdeploy_svc_blue)
                reqLogger.Info("Creating xds Service", "Xds.Namespace", blueSvc.Namespace, "Xds.Name", blueSvc.Name)
                err = r.client.Create(context.TODO(), blueSvc)
                if err != nil {
                    reqLogger.Error(err, "Failed to create new Service", "Deployment.Namespace", blueSvc.Namespace, "Deployment.Name", blueSvc.Name)
                    return reconcile.Result{}, err
                }
                // Deployment created successfully - return and requeue
                return reconcile.Result{Requeue: true}, nil
            } else if err != nil {
                reqLogger.Error(err, "Failed to get Deployment")
                return reconcile.Result{}, err
            }



            // Check if the green deployment already exists, if not create a new one
            err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_dep_green, Namespace: instance.Namespace}, depfound)
            if err != nil && errors.IsNotFound(err) {
                // Define a new deployment
                reqLogger.Info("Defining a new Deployment for: " + bgdeploy_dep_green)
                greenDep := r.newBGDeploymentForCR(instance, bgdeploy_dep_green, greenImage, l_bgdeploy_dep_green)
                reqLogger.Info("Creating a App Deployment", "Deployment.Namespace", greenDep.Namespace, "Deployment.Name", greenDep.Name)
                err = r.client.Create(context.TODO(), greenDep)
                if err != nil {
                    reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", greenDep.Namespace, "Deployment.Name", greenDep.Name)
                    return reconcile.Result{}, err
                }
                // Deployment created successfully - return and requeue
                return reconcile.Result{Requeue: true}, nil
            } else if err != nil {
                reqLogger.Error(err, "Failed to get Deployment")
                return reconcile.Result{}, err
            }


            // Check if the green service already exists, if not create a new one
            err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_svc_green, Namespace: instance.Namespace}, svcfound)
            if err != nil && errors.IsNotFound(err) {
                greenSvc := r.newBGServiceForCR(instance, bgdeploy_svc_green, l_bgdeploy_svc_green)
                reqLogger.Info("Creating xds Service", "Xds.Namespace", greenSvc.Namespace, "Xds.Name", greenSvc.Name)
                err = r.client.Create(context.TODO(), greenSvc)
                if err != nil {
                    reqLogger.Error(err, "Failed to create new Service", "Deployment.Namespace", greenSvc.Namespace, "Deployment.Name", greenSvc.Name)
                    return reconcile.Result{}, err
                }
                // Deployment created successfully - return and requeue
                return reconcile.Result{Requeue: true}, nil
            } else if err != nil {
                reqLogger.Error(err, "Failed to get Deployment")
                return reconcile.Result{}, err
            }


	// set the upstreamURL&Port to the Green Service
	reqURL := fmt.Sprintf("http://%s.%s:18080/xds?host=%s.%s&port=5000", bgdeploy_svc_xds, instance.Namespace, bgdeploy_svc_green, instance.Namespace)
        reqLogger.Info( "Changing upstreamURL: " + reqURL)
	if _, err := http.Get(reqURL); err != nil {
		reqLogger.Error( err, "failed to set the upstreamURL to Service: ", bgdeploy_svc_green)
		return reconcile.Result{}, err
	}




        } else if activeApp == "GREEN" && transitFlag == "OFF" {

// ELSE IF active==green && transit==OFF {
//      terminate bgdeploy-dep-blue  and bgdeploy-svc-blue  if they are not running there
//      bring-up  bgdeploy-dep-green and bgdeploy-svc-green if they are not running there
//      }

            // Check if the green deployment already exists, if not create a new one
            err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_dep_green, Namespace: instance.Namespace}, depfound)
            if err != nil && errors.IsNotFound(err) {
                // Define a new deployment
                reqLogger.Info("Defining a new Deployment for: " + bgdeploy_dep_green)
                greenDep := r.newBGDeploymentForCR(instance, bgdeploy_dep_green, greenImage, l_bgdeploy_dep_green)
                reqLogger.Info("Creating a App Deployment", "Deployment.Namespace", greenDep.Namespace, "Deployment.Name", greenDep.Name)
                err = r.client.Create(context.TODO(), greenDep)
                if err != nil {
                    reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", greenDep.Namespace, "Deployment.Name", greenDep.Name)
                    return reconcile.Result{}, err
                }
                // Deployment created successfully - return and requeue
                return reconcile.Result{Requeue: true}, nil
            } else if err != nil {
                reqLogger.Error(err, "Failed to get Deployment")
                return reconcile.Result{}, err
            }


            // Check if the green service already exists, if not create a new one
            err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_svc_green, Namespace: instance.Namespace}, svcfound)
            if err != nil && errors.IsNotFound(err) {
                greenSvc := r.newBGServiceForCR(instance, bgdeploy_svc_green, l_bgdeploy_svc_green)
                reqLogger.Info("Creating xds Service", "Xds.Namespace", greenSvc.Namespace, "Xds.Name", greenSvc.Name)
                err = r.client.Create(context.TODO(), greenSvc)
                if err != nil {
                    reqLogger.Error(err, "Failed to create new Service", "Deployment.Namespace", greenSvc.Namespace, "Deployment.Name", greenSvc.Name)
                    return reconcile.Result{}, err
                }
                // Deployment created successfully - return and requeue
                return reconcile.Result{Requeue: true}, nil
            } else if err != nil {
                reqLogger.Error(err, "Failed to get Deployment")
                return reconcile.Result{}, err
            }


            // Check if the blue deployment already exists, if yes delete that
            err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_dep_blue, Namespace: instance.Namespace}, depfound)
            if err != nil && errors.IsNotFound(err) {
                reqLogger.Info( "Blue Deployment not found")
            } else  {
                reqLogger.Info( "Deleting the Deployment")
                if err := r.client.Delete(context.TODO(), depfound); err != nil {
                    reqLogger.Error( err, "failed to delete Deployment resource")
                    return reconcile.Result{}, err
                }
                return reconcile.Result{}, err
            }


            // Check if the blue service already exists, if yes delete that
            err = r.client.Get(context.TODO(), types.NamespacedName{Name: bgdeploy_svc_blue, Namespace: instance.Namespace}, svcfound)
            if err != nil && errors.IsNotFound(err) {
                reqLogger.Info( "Blue Service not found")
                        } else  {
                reqLogger.Info( "Deleting the Service")
                if err := r.client.Delete(context.TODO(), svcfound); err != nil {
                    reqLogger.Error( err, "failed to delete Service resource")
                    return reconcile.Result{}, err
                }
                return reconcile.Result{}, err
            }


	// set the upstreamURL&Port to the Green Service
	reqURL := fmt.Sprintf("http://%s.%s:18080/xds?host=%s.%s&port=5000", bgdeploy_svc_xds, instance.Namespace, bgdeploy_svc_green, instance.Namespace)
        reqLogger.Info( "Changing upstreamURL: " + reqURL)
	if _, err := http.Get(reqURL); err != nil {
		reqLogger.Error( err, "failed to set the upstreamURL to Service: ", bgdeploy_svc_green)
		return reconcile.Result{}, err
	}



        }












        // Deployment and Service already exist - don't requeue
//        reqLogger.Info("Skip reconcile: Deployment and Service already exists", "Deployment.Name", depfound.Name, "Service.Name", svcfound.Name)
        return reconcile.Result{}, nil


        }




//     Create Blue or Green Deployment if it isn't there
// newBGDeploymentForCR returns a busybox pod with the same name/namespace as the cr
func (r *ReconcileBGDeploy) newBGDeploymentForCR(cr *swallowlabv1alpha1.BGDeploy, dep_name string, image_name string, bg_label map[string]string) *appsv1.Deployment {
    dep := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name: dep_name,
            Namespace: cr.Namespace,
            Labels: bg_label,
        },
        Spec: appsv1.DeploymentSpec{
            Selector: &metav1.LabelSelector{
                MatchLabels: bg_label,
            },
          Replicas: &cr.Spec.Replicas,
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{Labels: bg_label },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name: "bgdeploy",
                           Image: image_name,
                            Ports: []corev1.ContainerPort{{
//                                ContainerPort: &cr.Spec.Port,
                                ContainerPort: 5000,
                            }},
                            Env: []corev1.EnvVar{
                                {
                                    Name: "K8S_NODE_NAME",
                                    ValueFrom: &corev1.EnvVarSource{ FieldRef: &corev1.ObjectFieldSelector{ FieldPath: "spec.nodeName" }},
                              },
                                {
                                    Name: "K8S_POD_NAME",
                                    ValueFrom: &corev1.EnvVarSource{ FieldRef: &corev1.ObjectFieldSelector{ FieldPath: "metadata.name" }},
                                },
                                {
                                    Name: "K8S_POD_IP",
                                    ValueFrom: &corev1.EnvVarSource{ FieldRef: &corev1.ObjectFieldSelector{ FieldPath: "status.podIP" }},
                                },
                            },
                        },
                   },
                },
            },
        },
    }
    controllerutil.SetControllerReference(cr, dep, r.scheme)
    return dep
}


//     Create Blue or Green Service if it isn't there
func (r *ReconcileBGDeploy) newBGServiceForCR(cr *swallowlabv1alpha1.BGDeploy, svc_name string, bg_label map[string]string) *corev1.Service {
    svc := &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name: svc_name,
            Namespace: cr.Namespace,
        },
        Spec: corev1.ServiceSpec{
            Ports: []corev1.ServicePort{{
                Protocol: "TCP",
                Port: 5000,
                TargetPort: intstr.FromInt(5000),
            }},
            Type: corev1.ServiceTypeNodePort,
            Selector: bg_label,
        },
    }
    controllerutil.SetControllerReference(cr, svc, r.scheme)
    return svc
}




// If Envoy Pod is not there, create a new one
// newPodForCR returns a busybox pod with the same name/namespace as the cr
func  (r *ReconcileBGDeploy) newEnvoyPodForCR(cr *swallowlabv1alpha1.BGDeploy, pod_name string, pod_label map[string]string) *corev1.Pod {

        configmapvolume := &corev1.ConfigMapVolumeSource{
                        LocalObjectReference: corev1.LocalObjectReference{Name: pod_name + "-configmap"},
        }
        pod := &corev1.Pod{
                        ObjectMeta: metav1.ObjectMeta{
                                        Name:      pod_name,
                                        Namespace: cr.Namespace,
                                        Labels:    pod_label,
                        },
                        Spec: corev1.PodSpec{
                                        Containers: []corev1.Container{{
                                                                        Name:    "envoy",
                                                                        Image:   "envoyproxy/envoy:v1.15-latest",
                                                                        Command: []string{"/usr/local/bin/envoy"},
                                                                        Args:    []string{"--config-path /etc/envoy/envoy.yaml"},
                                                                        VolumeMounts: []corev1.VolumeMount{{
                                                                                Name: "envoy",
                                                                                MountPath: "/etc/envoy",
                                                                        }},
                                                                        Ports: []corev1.ContainerPort{{
                                                                                ContainerPort: 10000,
                                                                        }},
                                        }},
                                        Volumes:  []corev1.Volume{{
                                                                        Name: "envoy",
                                                                        VolumeSource: corev1.VolumeSource{
                                                                                ConfigMap: configmapvolume,
                                                                        },
                                        }},
                        },
        }
        controllerutil.SetControllerReference(cr, pod, r.scheme)
        return pod
}



// If Envoy Service is not there, create a new one
func (r *ReconcileBGDeploy) newEnvoyServiceForCR(cr *swallowlabv1alpha1.BGDeploy, svc_name string, svc_label map[string]string) *corev1.Service {

        svc := &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name: svc_name,
            Namespace: cr.Namespace,
        },
        Spec: corev1.ServiceSpec{
            Ports: []corev1.ServicePort{{
                Protocol: "TCP",
                Port: 10000,
                TargetPort: intstr.FromInt(10000),
            }},
            Type: corev1.ServiceTypeNodePort,
            Selector: svc_label,
        },
    }
    controllerutil.SetControllerReference(cr, svc, r.scheme)
    return svc
}



// If XDS Service is not there, create a new one
func (r *ReconcileBGDeploy) newXDSServiceForCR(cr *swallowlabv1alpha1.BGDeploy, svc_name string, svc_label map[string]string) *corev1.Service {
/*
        xdsports1 := corev1.ServicePort{
                {
                Name: "xds-grpc",
                Protocol: "TCP",
                Port: 18000,
                TargetPort: intstr.FromInt(18000),
        }}
*/
        xdsport1 := corev1.ServicePort{}
        xdsport1.Name = "xds-grpc"
        xdsport1.Protocol = "TCP"
        xdsport1.Port = 18000
        xdsport1.TargetPort = intstr.FromInt(18000)


        xdsport2 := corev1.ServicePort{}
        xdsport2.Name = "snapshot-http"
        xdsport2.Protocol = "TCP"
        xdsport2.Port = 18080
        xdsport2.TargetPort = intstr.FromInt(18080)

/*
        xdsports2 := corev1.ServicePort{{
                Name: "snapshot-http",
                Protocol: "TCP",
                Port: 18080,
                TargetPort: intstr.FromInt(18080)}}

    //    var xdsports [2]corev1.ServicePort = {xdsport1, xdsport2}
        xdsports :=[]corev1.ServicePort{ xdsport1 }
        xdsports =append(xdsports, xdsport2 )

        xdsports := []corev1.ServicePort{}
        xdsports[0] = xdsport1
        xdsports[1] = xdsport2
*/
        xdsports :=[]corev1.ServicePort{ xdsport1 }
        xdsports =append(xdsports, xdsport2 )






        svc := &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name: svc_name,
            Namespace: cr.Namespace,
        },
        Spec: corev1.ServiceSpec{
            Ports: xdsports,
            Type: corev1.ServiceTypeNodePort,
//            Type: corev1.ServiceTypeClusterIP,
            Selector: svc_label,
        },
    }
    controllerutil.SetControllerReference(cr, svc, r.scheme)
    return svc
}






// newLabelsForCR returns the labels for selecting the resources
// belonging to the given CR name.
func newLabelsForCR(name string) map[string]string {
    return map[string]string{"app": "bgdeploy", "bgdeploy_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
    var podNames []string
    for _, pod := range pods {
        podNames = append(podNames, pod.Name)
    }
    return podNames
}


// RECONCILE BLOCK END


