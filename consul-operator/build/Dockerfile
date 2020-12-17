FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/consul-operator \
    USER_UID=1001 \
    USER_NAME=consul-operator \
    RESREQ_DIR=/usr/src/app/resource-reqs-generated \
    DEPLOYMENT_DIR=/usr/src/app


COPY deployment/helm /usr/local/bin/helm
RUN chmod +x  /usr/local/bin/helm
COPY deployment/app-deployment /usr/src/app/app-deployment
COPY deployment/resource-reqs /usr/src/app/resource-reqs

RUN chmod -R 770 /usr/src/app/

# install operator binary
COPY build/_output/bin/consul-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  chmod u+x /usr/local/bin/user_setup
RUN  chmod u+x /usr/local/bin/entrypoint

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
