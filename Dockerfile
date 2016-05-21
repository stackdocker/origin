#
# This is the unofficial OpenShift Origin image for the DockerHub. It has as its
# entrypoint the OpenShift all-in-one binary.
#
# See images/origin for the official release version of this image
#
# The standard name for this image is openshift/origin
#
FROM openshift/origin-base

# by tangfx
#RUN yum install -y golang && yum clean all

#WORKDIR /go/src/github.com/openshift/origin
#ADD .   /go/src/github.com/openshift/origin
#ENV GOPATH /go
#ENV PATH $PATH:$GOROOT/bin:$GOPATH/bin
#
#RUN go get github.com/openshift/origin && \
#    hack/build-go.sh && \
#    cp _output/local/bin/linux/amd64/* /usr/bin/ && \
#    mkdir -p /var/lib/origin
COPY openshift oc /usr/bin/
WORKDIR /usr/bin
RUN  mkdir -p /var/lib/origin && \
     ln -s openshift atomic-enterprise && ln -s openshift kubectl && ln -s openshift kubernetes && \
     ln -s openshift openshift-docker-build && ln -s openshift openshift-sti-build && ln -s openshift osc && \
     ln -s openshift kube-apiserver && ln -s openshift kubelet && ln -s openshift kube-scheduler && \
     ln -s openshift openshift-recycle && ln -s openshift origin && \
     ln -s openshift kube-controller-manager && ln -s openshift kube-proxy && ln -s openshift oadm && \
     ln -s openshift openshift-deploy && ln -s openshift openshift-router && ln -s openshift osadm

EXPOSE 8080 8443
WORKDIR /var/lib/origin
ENTRYPOINT ["/usr/bin/openshift"]
