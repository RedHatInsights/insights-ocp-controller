FROM centos:centos7

MAINTAINER Lindani Phiri <lphiri@redhat.com>

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

RUN go build -o insights-controller

ENTRYPOINT ["./insights-controller"]
