default: build

.PHONY: build
build:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o terraform-provider-acecloud

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/akshay-malik-rtds/acecloud/0.1.0/$$(go env GOOS)_$$(go env GOARCH)
	mv terraform-provider-acecloud ~/.terraform.d/plugins/registry.terraform.io/akshay-malik-rtds/acecloud/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/

.PHONY: test
test:
	CGO_ENABLED=0 go test -v -cover -timeout=120s -parallel=4 ./...

.PHONY: testacc
testacc:
	TF_ACC=1 CGO_ENABLED=0 go test -v -cover -timeout=120m ./...

.PHONY: fmt
fmt:
	gofmt -s -w -e .

.PHONY: lint
lint:
	golangci-lint run ./... --timeout=5m

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: docs
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate \
		--provider-dir . \
		--provider-name acecloud \
		--rendered-provider-name "AceCloud"

.PHONY: docs-validate
docs-validate:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs validate \
		--provider-dir . \
		--provider-name acecloud

.PHONY: clean
clean:
	rm -f terraform-provider-acecloud terraform-provider-acecloud_v*
	rm -rf dist/

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         Build the provider binary"
	@echo "  install       Build and install locally for terraform dev_overrides"
	@echo "  test          Run unit tests"
	@echo "  testacc       Run acceptance tests (requires TF_ACC=1, real credentials)"
	@echo "  fmt           Format source code"
	@echo "  lint          Run golangci-lint"
	@echo "  vet           Run go vet"
	@echo "  tidy          Tidy go.mod"
	@echo "  docs          Generate provider documentation via tfplugindocs"
	@echo "  docs-validate Validate generated docs"
	@echo "  clean         Remove built binaries and dist/"
