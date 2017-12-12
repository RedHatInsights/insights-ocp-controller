FROM centos:centos7
MAINTAINER Red Hat Insights Team

COPY ./build/cmd /insights_controller

ENTRYPOINT ["/insights_controller"]
