FROM golang:1.18-alpine AS build

RUN apk update && \
    apk add --no-cache \
        git \
        musl-dev \
        build-base \
        git \
        openssh \
        ca-certificates \
        pkgconf \
	bash \
	vim \
	curl

RUN mkdir /app
WORKDIR /app

ADD . .

ENV GOOS=linux
ENV GOARCH=amd64
RUN go mod download
RUN go build

ENTRYPOINT ["/kafka-producer-benchmark"]
