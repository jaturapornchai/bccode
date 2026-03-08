# syntax=docker/dockerfile:1

##
## Builder
##
FROM golang:1.26-alpine AS builder

RUN apk add alpine-sdk
RUN apk add librdkafka-dev
RUN apk add build-base

WORKDIR /go/app

COPY go.mod /go/app
COPY go.sum /go/app
RUN go mod download

ADD . /go/app

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o go-app -tags musl main.go

##
## Deploy
##
FROM alpine:3.21

RUN apk add --no-cache librdkafka ca-certificates && update-ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app/
COPY --from=builder /go/app/go-app .
COPY ./tdict-std.txt /app/tdict-std.txt
COPY ./assets/fonts/ /app/fonts/
COPY ./assets/language/ /app/language/

RUN chown -R appuser:appgroup /app

ENV DEV_API_MODE=2
ENV SERVICE_PORT=8888
ENV TZ=Asia/Bangkok

USER appuser

EXPOSE 8888

HEALTHCHECK --interval=30s --timeout=10s --start-period=120s --retries=5 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8888/healthz || exit 1

ENTRYPOINT ["/app/go-app"]
