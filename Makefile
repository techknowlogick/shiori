DIST := release
IMPORT := src.techknowlogick.com/shiori
GO ?= go
SED_INPLACE := sed -i
GOFILES := $(shell find . -name "*.go" -type f ! -path "./vendor/*" ! -path "*/*-packr.go")
GOFMT ?= gofmt -s
SHASUM := shasum -a 256
SHELL := bash

TAGS ?=
LDFLAGS ?=

# $(call strip-suffix,filename)
strip-suffix = $(firstword $(subst ., ,$(1)))

.PHONY: fmt
fmt:
	$(GOFMT) -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	# get all go files and run go fmt on them
	@diff=$$($(GOFMT) -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

# dist step is kept for backwords compatibility
.PHONY: dist
dist: dep-node dep-go

.PHONY: dep
dep: dep-node dep-go

.PHONY: dep-node
dist-node:
	@hash npx > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		echo "Please install npm version 5.2+"; \
		exit 1; \
	fi;
	npm install
	npx parcel build src/*.html --public-url /dist/

.PHONY: dep-go
dist-go:
	@hash packr2 > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/gobuffalo/packr/v2/packr2; \
	fi
	packr2

.PHONY: release
release: release-windows release-darwin release-linux release-compress release-check

.PHONY: release-windows
release-windows:
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/techknowlogick/xgo; \
	fi
	go get -u github.com/mattn/go-isatty # needed for progress bar in windows
	go get -u github.com/inconshreveable/mousetrap # needed for windows builds
	mkdir -p "$(GOPATH)/src/github.com/konsorten"
	git clone https://github.com/konsorten/go-windows-terminal-sequences.git "$(GOPATH)/src/github.com/konsorten/go-windows-terminal-sequences"
	xgo -dest $(DIST) -tags 'netgo $(TAGS)' -ldflags '-linkmode external -extldflags "-static" $(LDFLAGS)' -targets 'windows/*' -out shiori .
ifeq ($(CI),drone)
	mv /build/* $(DIST)/
endif

.PHONY: release-darwin
release-darwin:
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/techknowlogick/xgo; \
	fi
	xgo -dest $(DIST) -tags 'netgo $(TAGS)' -ldflags '$(LDFLAGS)' -targets 'darwin/*' -out shiori .
ifeq ($(CI),drone)
	mv /build/* $(DIST)/
endif

.PHONY: release-linux
release-linux:
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/techknowlogick/xgo; \
	fi
	xgo -dest $(DIST) -tags 'netgo $(TAGS)' -ldflags '-linkmode external -extldflags "-static" $(LDFLAGS)' -targets 'linux/*' -out shiori .
ifeq ($(CI),drone)
	mv /build/* $(DIST)/
endif

.PHONY: release-check
release-check:
	cd $(DIST); for file in `find . -type f -name "*"`; do $(SHASUM) $${file:2} > $${file}.sha256; done;

.PHONY: release-compress
release-compress:
	@hash gxz > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/ulikunitz/xz/cmd/gxz; \
	fi
	cd $(DIST); for file in `find . -type f -name "*"`; do echo "compressing $${file}" && gxz -k -9 $${file}; done;
