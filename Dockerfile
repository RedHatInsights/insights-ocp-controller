FROM centos:centos7

MAINTAINER Jeremy Crafts <jcrafts@redhat.com>

LABEL com.redhat.component="insights-ocp-controller-container"
LABEL name="containers/insights-ocp-controller"
LABEL version="0.1"
LABEL summary="Controller container for Red Hat Insights on Openshift"

COPY . /go/src/github.com/RedHatInsights/insights-ocp-controller
ENV GOPATH=/go

WORKDIR /go/src/github.com/RedHatInsights/insights-ocp-controller

RUN yum install -y epel-release --enablerepo=extras
RUN yum install --enablerepo=epel-testing -y golang insights-client-3.0.3-2.fc27.noarch.rpm && \
   go get k8s.io/client-go/... && go build -o /usr/bin/insights-ocp-controller 


ENV EGG=/etc/insights-client/rpm.egg

ENTRYPOINT ["/usr/bin/insights-ocp-controller"]
