
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

.PHONY: dist
dist:
	@hash packr2 > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/gobuffalo/packr/v2/packr2; \
	fi
	npm install
	npx parcel build src/*.html --public-url /dist/
	packr2
