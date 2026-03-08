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

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build  -o go-app -tags musl ./cmd/app/main.go  

##
## Deploy
##
FROM alpine:latest 
WORKDIR /root/
COPY --from=builder /go/app/go-app .
# ADD ./cmd/app/.env /root/
ENTRYPOINT  /root/go-app