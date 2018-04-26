FROM centos:centos7

MAINTAINER Jeremy Crafts <jcrafts@redhat.com>

LABEL com.redhat.component="insights-ocp-controller-container"
LABEL name="containers/insights-ocp-controller"
LABEL version="0.1"
LABEL summary="Controller container for Red Hat Insights on Openshift"

RUN yum -y update-minimal --security --sec-severity=Important --sec-severity=Critical --setopt=tsflags=nodocs && \
    yum install -y golang git && \
    yum clean all

ENV GOPATH=/go
RUN mkdir -p /go/src/github.com/RedHatInsights/insights-ocp/controller
WORKDIR /go/src/github.com/RedHatInsights/insights-ocp/controller

COPY cmd.go .
COPY pkg ./pkg
COPY vendor ./vendor
RUN ls

RUN go build -o /insights-controller

ENTRYPOINT ["./insights-controller"]