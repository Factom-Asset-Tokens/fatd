FROM golang:1.13-alpine AS builder

RUN apk add make bash git gcc libc-dev

ARG GOBIN=/go/bin/
ARG GOOS=linux
ARG GOARCH=amd64
ARG GO111MODULE=on
ARG PKG_NAME=github.com/Factom-Asset-Tokens/fatd
ARG PKG_PATH=${GOPATH}/src/${PKG_NAME}

WORKDIR ${PKG_PATH}
COPY . ${PKG_PATH}/
RUN make

FROM alpine:3.10

COPY --from=builder /go/src/github.com/Factom-Asset-Tokens/fatd/fatd .

ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.2.1/wait /wait
RUN chmod +x /wait

ENTRYPOINT [ "./fatd" ]
