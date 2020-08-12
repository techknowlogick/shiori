FROM node:lts-alpine as nodebuilder
RUN apk --no-cache add python2 make bash git

WORKDIR /app

COPY . .

RUN make dep-node

FROM golang:1.15-alpine as gobuilder

RUN apk update \
  && apk --no-cache add git build-base make bash

WORKDIR /go/src/src.techknowlogick.com/shiori
COPY . .
COPY --from=nodebuilder /app/dist /go/src/src.techknowlogick.com/shiori/dist/
RUN go get -u github.com/markbates/pkger/cmd/pkger
RUN pkger && make build

FROM alpine:3.12

ENV ENV_SHIORI_DIR /srv/shiori/

RUN apk --no-cache add dumb-init ca-certificates
COPY --from=gobuilder /go/src/src.techknowlogick.com/shiori/shiori /usr/local/bin/shiori
COPY --from=gobuilder /go/src/src.techknowlogick.com/shiori/dist /dist

WORKDIR /srv/
RUN mkdir shiori

EXPOSE 8080

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/usr/local/bin/shiori", "serve"]
