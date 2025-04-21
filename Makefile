.PHONY: run build test coverage lint clean check fmt vet install-tools

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOLINT=golangci-lint
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Binary name
BINARY_NAME=wachturm
CMD_DIR=./cmd/wachturm

all: check test build

run:
	$(GOCMD) run $(CMD_DIR)

build:
	$(GOBUILD) -o ./dist/$(BINARY_NAME) -v $(CMD_DIR)

test:
	$(GOTEST) -v ./...

coverage:
	$(GOTEST) -coverprofile=coverage.out -coverpkg=./... ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

coverage-text:
	$(GOTEST) -coverprofile=coverage.out -coverpkg=./... ./...
	$(GOCMD) tool cover -func=coverage.out

lint:
	$(GOLINT) run

fmt:
	$(GOFMT) ./...

vet:
	$(GOVET) ./...

check: fmt vet lint

clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	rm -f coverage.html

tidy:
	$(GOMOD) tidy

install-tools:
	@echo "Installing golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.56.2
	@echo "Tools installed successfully"