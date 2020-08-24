FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/bgdeploy-xds2 \
    USER_UID=1001 \
    USER_NAME=bgdeploy-xds2

# install operator binary
COPY build/_output/bin/bgdeploy-xds2 ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
