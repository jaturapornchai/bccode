# syntax=docker/dockerfile:1

##
## Builder
##
FROM golang:1.17-alpine3.15 AS builder

RUN apk add alpine-sdk
RUN apk add librdkafka=1.8.2-r0
RUN apk add build-base

WORKDIR /go/app

COPY go.mod /go/app
COPY go.sum /go/app
RUN go mod download

ADD . /go/app

RUN find . -name "*_test.go" -exec rm -rf {} \;

ENV ENVIRONMENT=production

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build  -o go-app -tags musl ./cmd/authenticationservice/main.go  

##
## Deploy
##
FROM alpine:latest 
WORKDIR /root/
ENV SERVICE_NAME authenticaton-api

COPY --from=builder /go/app/go-app .
ENTRYPOINT  /root/go-app