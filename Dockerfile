FROM golang:1.5.2
MAINTAINER colin.hom@coreos.com

RUN go get github.com/tools/godep

ADD . $GOPATH/src/github.com/coreos/khealth

WORKDIR $GOPATH/src/github.com/coreos/khealth

RUN godep go install github.com/coreos/khealth/cmd/rcscheduler

EXPOSE 8080
