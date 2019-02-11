FROM node:lts-alpine as nodebuilder
RUN apk --no-cache add python2

WORKDIR /app

COPY . .

RUN npm install && npx parcel build src/*.html --public-url /dist/ 

FROM golang:1.11-alpine as gobase

RUN apk update \
  && apk --no-cache add git build-base

ENV GO111MODULE=on
WORKDIR /go/src/github.com/techknowlogick/shiori

COPY . .
RUN go mod download && go mod vendor

FROM golang:1.11-alpine as gobuilder

RUN apk update \
  && apk --no-cache add git build-base

WORKDIR /go/src/github.com/techknowlogick/shiori

COPY . .
COPY --from=gobase /go/src/github.com/techknowlogick/shiori/vendor /go/src/github.com/techknowlogick/shiori/vendor/
COPY --from=nodebuilder /app/dist /go/src/github.com/techknowlogick/shiori/dist/
RUN go get -u github.com/gobuffalo/packr/v2/packr2
RUN packr2 build -mod vendor -o shiori

FROM alpine:3.9

ENV ENV_SHIORI_DIR /srv/shiori/

RUN apk --no-cache add dumb-init ca-certificates
COPY --from=gobuilder /go/src/github.com/techknowlogick/shiori/shiori /usr/local/bin/shiori

WORKDIR /srv/
RUN mkdir shiori

EXPOSE 8080

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/usr/local/bin/shiori", "serve"]
