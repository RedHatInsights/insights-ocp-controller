FROM centos:centos7

MAINTAINER Jeremy Crafts <jcrafts@redhat.com>

LABEL com.redhat.component="insights-ocp-controller-container"
LABEL name="containers/insights-ocp-controller"
LABEL version="0.1"
LABEL summary="Controller container for Red Hat Insights on Openshift"

COPY . /go/src/github.com/RedHatInsights/insights-ocp-controller
ENV GOPATH=/go

WORKDIR /go/src/github.com/RedHatInsights/insights-ocp-controller

RUN yum install -y golang git && \
   go build -o /insights-controller

ENTRYPOINT ["./insights-controller"]