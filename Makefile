TARGETOS ?= linux
TARGETARCH ?= amd64
LDFLAGS := -s -w -extldflags "-static"

INSECURE_TAG :=
ifeq ($(INSECURE),true)
    INSECURE_TAG += insecure
endif

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags '${LDFLAGS}' -v -a -tags netgo,${INSECURE_TAG} -o plugin-docker-buildx ./cmd/docker-buildx

.PHONY: format
format: install-tools
	gofumpt -extra -w .

.PHONY: formatcheck
formatcheck: install-tools
	@([ -z "$(shell gofumpt -d . | head)" ]) || (echo "Source is unformatted"; exit 1)

.PHONY: install-tools
install-tools: ## Install development tools
	@hash gofumpt > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install mvdan.cc/gofumpt@latest; \
	fi

.PHONY: test
test:
	go test -tags netgo,${INSECURE_TAG} -cover ./...
