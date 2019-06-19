FROM node:lts-alpine as nodebuilder
RUN apk --no-cache add python2 make bash git

WORKDIR /app

COPY . .

RUN make dep-node

FROM golang:1.12-alpine as gobuilder

RUN apk update \
  && apk --no-cache add git build-base make bash

WORKDIR /go/src/src.techknowlogick.com/shiori
COPY . .
ENV GO111MODULE=on
RUN go mod download && go mod vendor
COPY --from=nodebuilder /app/dist /go/src/src.techknowlogick.com/shiori/dist/
RUN GO111MODULE=auto go get -u github.com/gobuffalo/packr/v2/packr2
RUN packr2 && make build

FROM alpine:3.10

ENV ENV_SHIORI_DIR /srv/shiori/

RUN apk --no-cache add dumb-init ca-certificates
COPY --from=gobuilder /go/src/src.techknowlogick.com/shiori/shiori /usr/local/bin/shiori

WORKDIR /srv/
RUN mkdir shiori

EXPOSE 8080

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/usr/local/bin/shiori", "serve"]
