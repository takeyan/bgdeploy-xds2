package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BGDeploySpec defines the desired state of BGDeploy
type BGDeploySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
            Blue      string `json:"blue"`
            Green     string `json:"green"`
            Port      int32 `json:"port"`
            Replicas  int32 `json:"replicas"`       // Deploymentマニフェストのreplicasに反映
            Transit   string `json:"transit"`
            Active    string `json:"active"`
}

// BGDeployStatus defines the observed state of BGDeploy
type BGDeployStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
            Nodes []string `json:"nodes"`  // デプロイされたPod名一覧を保持
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BGDeploy is the Schema for the bgdeploys API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=bgdeploys,scope=Namespaced
type BGDeploy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BGDeploySpec   `json:"spec,omitempty"`
	Status BGDeployStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BGDeployList contains a list of BGDeploy
type BGDeployList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BGDeploy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BGDeploy{}, &BGDeployList{})
}
