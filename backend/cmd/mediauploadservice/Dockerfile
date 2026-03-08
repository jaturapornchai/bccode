
## smlsoft/smlcloudplatform:shop
##
## Builder
##
FROM golang:1.20.2-alpine3.17 AS builder

RUN apk add alpine-sdk
RUN apk add librdkafka=1.8.2-r0
RUN apk add build-base

WORKDIR /go/app

COPY go.mod /go/app
COPY go.sum /go/app
RUN go mod download

ADD . /go/app

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build  -o go-app -tags musl ./cmd/mediauploadservice/main.go  

##
## Deploy
##
FROM alpine:latest 
WORKDIR /root/
# ADD ./cmd/authenticationservice/.env .
ENV SERVICE_NAME media-upload-api

COPY --from=builder /go/app/go-app .
ENTRYPOINT  /root/go-app