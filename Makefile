BINARY := governance-audit
MODULE := github.com/MightyDestroyer/Governance/tools/governance-audit
VERSION := 1.0.0

GOFLAGS := -trimpath
LDFLAGS := -s -w

.PHONY: build test lint fmt clean install help

build:
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) .

test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

clean:
	rm -f $(BINARY)
	rm -f $(BINARY).exe
	go clean

install:
	go install $(GOFLAGS) -ldflags "$(LDFLAGS)" .

help:
	@echo "Governance Audit CLI — Build Targets"
	@echo ""
	@echo "  make build     Build the binary"
	@echo "  make test      Run tests with race detector"
	@echo "  make lint      Run golangci-lint"
	@echo "  make fmt       Format code (gofmt + goimports)"
	@echo "  make clean     Remove build artifacts"
	@echo "  make install   Install via go install"
	@echo "  make help      Show this help"
