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
	curl

RUN mkdir /app
WORKDIR /app

ADD . .

ENV GOOS=linux
ENV GOARCH=amd64
RUN go mod download
RUN go build


FROM alpine AS release
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/kafka-producer-benchmark /
ENTRYPOINT ["/kafka-producer-benchmark"]
