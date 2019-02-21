DIST := .
IMPORT := src.techknowlogick.com/shiori
GO ?= go
SED_INPLACE := sed -i
GOFILES := $(shell find . -name "*.go" -type f ! -path "./vendor/*" ! -path "*/*-packr.go")
GOFMT ?= gofmt -s

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
release: cross release-compress release-check

.PHONY: cross
cross:
	@hash gox > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/mitchellh/gox; \
	fi
	go get -u github.com/mattn/go-isatty # needed for progress bar in windows
	go get -u github.com/inconshreveable/mousetrap # needed for windows builds
	mkdir -p "$(GOPATH)/src/github.com/konsorten"
	git clone https://github.com/konsorten/go-windows-terminal-sequences.git "$(GOPATH)/src/github.com/konsorten/go-windows-terminal-sequences"
	gox -output "release/shiori_{{.OS}}_{{.Arch}}" -ldflags "-X main.version=`git rev-parse --short HEAD`" -verbose ./...

.PHONY: release-check
release-check:
	cd $(DIST)/release; $(foreach file,$(wildcard $(DIST)/release/$(EXECUTABLE)-*),sha256sum $(notdir $(file)) > $(notdir $(file)).sha256;)

.PHONY: release-compress
release-compress:
	@hash gxz > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/ulikunitz/xz/cmd/gxz; \
	fi
	cd $(DIST)/release; $(foreach file,$(wildcard $(DIST)/binaries/$(EXECUTABLE)-*),gxz -k -9 $(notdir $(file));)