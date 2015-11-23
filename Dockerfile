FROM golang:1.5
MAINTAINER colin.hom@coreos.com

RUN go get github.com/tools/godep

ADD . /go/src/github.com/coreos/khealth

WORKDIR /go/src/github.com/coreos/khealth

RUN godep restore

RUN go install github.com/coreos/khealth/cmd/rcscheduler

EXPOSE 8080



