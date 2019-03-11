DIST := dist
IMPORT := src.techknowlogick.com/shiori
GO ?= go
SED_INPLACE := sed -i
GOFILES := $(shell find . -name "*.go" -type f ! -path "./vendor/*" ! -path "*/*-packr.go")
GOFMT ?= gofmt -s
SHASUM := shasum -a 256
SHELL := bash

TAGS ?=

ifneq ($(DRONE_TAG),)
	VERSION ?= $(subst v,,$(DRONE_TAG))
	SHIORI_VERSION := $(VERSION)
else
	ifneq ($(DRONE_BRANCH),)
		VERSION ?= $(subst release/v,,$(DRONE_BRANCH))
	else
		VERSION ?= master
	endif
	SHIORI_VERSION ?= $(shell git describe --tags --always | sed 's/-/+/' | sed 's/^v//')
endif

LDFLAGS := -X "main.Version=$(SHIORI_VERSION)" -X "main.Tags=$(TAGS)"

ifeq ($(OS), Windows_NT)
	EXECUTABLE := shiori.exe
else
	EXECUTABLE := shiori
endif

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

.PHONY: build
build: $(EXECUTABLE)

$(EXECUTABLE): $(SOURCES)
	$(GO) build $(GOFLAGS) $(EXTRA_GOFLAGS) -tags '$(TAGS)' -ldflags '-s -w $(LDFLAGS)' -o $@

# dist step is kept for backwords compatibility
.PHONY: dist
dist: dep-node dep-go

.PHONY: dep
dep: dep-node dep-go

.PHONY: dep-node
dep-node:
	@hash npx > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		echo "Please install npm version 5.2+"; \
		exit 1; \
	fi;
	npm install
	npx parcel build src/*.html --public-url /dist/
	npx replace-x '(\.\./){1,3}shiori' '/dist' ./dist/ --include="*.css" -q -r

.PHONY: dep-go
dep-go:
	@hash packr2 > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/gobuffalo/packr/v2/packr2; \
	fi
	packr2

.PHONY: cross
cross: release-dirs release-windows release-darwin release-linux release-copy

.PHONY: release
release: release-compress release-check

.PHONY: release-dirs
release-dirs:
	mkdir -p $(DIST)/binaries $(DIST)/release

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
	cp /build/* $(DIST)/binaries
endif

.PHONY: release-darwin
release-darwin:
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/techknowlogick/xgo; \
	fi
	xgo -dest $(DIST) -tags 'netgo $(TAGS)' -ldflags '$(LDFLAGS)' -targets 'darwin/*' -out shiori .
ifeq ($(CI),drone)
	cp /build/* $(DIST)/binaries
endif

.PHONY: release-linux
release-linux:
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/techknowlogick/xgo; \
	fi
	xgo -dest $(DIST) -tags 'netgo $(TAGS)' -ldflags '-linkmode external -extldflags "-static" $(LDFLAGS)' -targets 'linux/*' -out shiori .
ifeq ($(CI),drone)
	cp /build/* $(DIST)/binaries
endif

.PHONY: release-copy
release-copy:
	cd $(DIST); for file in `find /build -type f -name "*"`; do cp $${file} ./release/; done;

.PHONY: release-check
release-check:
	cd $(DIST)/release/; for file in `find . -type f -name "*"`; do $(SHASUM) $${file:2} > $${file}.sha256; done;

.PHONY: release-compress
release-compress:
	@hash gxz > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/ulikunitz/xz/cmd/gxz; \
	fi
	cd $(DIST)/release/; for file in `find . -type f -name "*"`; do echo "compressing $${file}" && gxz -k -9 $${file}; done;
