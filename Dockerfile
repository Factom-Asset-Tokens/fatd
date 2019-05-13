FROM golang:1.12-alpine AS builder

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

FROM alpine:3.9

COPY --from=builder /go/src/github.com/Factom-Asset-Tokens/fatd/fatd .

ENTRYPOINT [ "./fatd" ]
