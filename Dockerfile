FROM golang:1.11-alpine as builder

RUN apk update \
  && apk --no-cache add git build-base

WORKDIR /go/src/github.com/techknowlogick/shiori
COPY . .
RUN go get -d -v ./...
RUN go build -o shiori

FROM alpine:3.9

ENV ENV_SHIORI_DIR /srv/shiori/

RUN apk --no-cache add dumb-init ca-certificates
COPY --from=builder /go/src/github.com/techknowlogick/shiori/shiori /usr/local/bin/shiori

WORKDIR /srv/
RUN mkdir shiori

EXPOSE 8080

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/usr/local/bin/shiori", "serve"]
