FROM golang:1.10-alpine as builder

RUN apk update \
  && apk --no-cache add git build-base

WORKDIR /go/src/code.techknowlogick.com/techknowlogick/shiori
COPY . .
RUN go get -d -v ./...
RUN go build -o shiori

FROM alpine:latest

ENV ENV_SHIORI_DIR /srv/shiori/

RUN apk --no-cache add dumb-init ca-certificates
COPY --from=builder /go/src/code.techknowlogick.com/techknowlogick/shiori/shiori /usr/local/bin/shiori

WORKDIR /srv/
RUN mkdir shiori

EXPOSE 8080

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/usr/local/bin/shiori", "serve"]