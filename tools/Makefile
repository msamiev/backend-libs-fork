export GOTOOLDIR := $(PWD)/tools/bin
export GOBIN     := $(GOTOOLDIR)
SHELL            := env PATH=$(GOTOOLDIR):$(PATH) $(SHELL) # https://stackoverflow.com/a/36226832
BELIBURL         := https://$(GH_USER):$(GH_TOKEN)@raw.githubusercontent.com/fusionmedialimited/backend-infra-libraries/master

GOLANGCI_LINT_VERSION := v1.52.2
MARKDOWNFMT_VERSION   := v3.1.0

GOFLAGS := -buildvcs=false

.PHONY: tools
tools:
	mkdir -p $(GOTOOLDIR)
	curl -sSfL $(BELIBURL)/scripts/golangci-lint-install.sh | VERSION=$(GOLANGCI_LINT_VERSION) sh -s
	go install github.com/Kunde21/markdownfmt/v3/cmd/markdownfmt@$(MARKDOWNFMT_VERSION)

.PHONY: lint
lint:
	golangci-lint run
	markdownfmt -gofmt -soft-wraps -l $$(find $(PWD) -type f -name "*.md" -not -path "$(PWD)/vendor/*") 2>&1 | \
		xargs test -z || (echo "Please run 'make fmt' to fix it issue" && exit 1)

.PHONY: fmt
fmt:
	golangci-lint run --fix --fast
	markdownfmt -gofmt -soft-wraps -w $$(find $(PWD) -type f -name "*.md" -not -path "$(PWD)/vendor/*")

.PHONY: generate
generate:
	find $(PWD) \( -type d -iname mocks \) -o \( -type f -iname "*.gen.go" \) -exec rm -rf {} \; || true
	go generate ./...
