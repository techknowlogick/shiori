
GOFILES := $(shell find . -name "*.go" -type f ! -path "./vendor/*" ! -path "*/assets-prod.go")
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

.PHONY: less-generate
less-generate:
	$(foreach file, $(filter-out public/less/variable.less, $(wildcard view/less/*)),node_modules/.bin/lessc --clean-css view/less/$(notdir $(file)) > view/css/$(notdir $(call strip-suffix,$(file))).css;)

.PHONY: less-check
less-check: less-generate
	@diff=$$(git diff view/css/*); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make less-generate' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
fi;
