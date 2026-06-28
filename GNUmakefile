default: build

BINARY  = terraform-provider-polymarket
VERSION ?= dev

.PHONY: build
build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) .

.PHONY: install
install:
	go install -ldflags "-X main.version=$(VERSION)" .

.PHONY: test
test:
	go test ./... -timeout=120s

# Acceptance tests hit the live Polymarket API; gate them behind TF_ACC.
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v -timeout=120m

.PHONY: fmt
fmt:
	gofmt -s -w .

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	go vet ./...
	gofmt -l .

# Regenerate registry docs from schema descriptions and examples.
.PHONY: docs
docs:
	go generate ./...

.PHONY: tidy
tidy:
	go mod tidy
