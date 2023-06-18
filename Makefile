GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOBUILDRACE=$(GOCMD) build -race
GOINSTALL=$(GOCMD) install
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt

BIN_NAME=merkle-tree

ifeq ($(GIT_SHA),)
GIT_SHA:=$(shell git rev-parse HEAD)
endif

LDFLAGS = "-X 'main.GitSHA1=$(GIT_SHA)'"

.PHONY: all test coverage build checkfmt fmt
all: test coverage build checkfmt fmt

build:
	$(GOBUILD) \
        -ldflags=$(LDFLAGS) -o $(BIN_NAME) main.go

build-race:
	$(GOBUILDRACE) \
        -ldflags=$(LDFLAGS) .

install:
	$(GOINSTALL)

checkfmt:
	@echo 'Checking gofmt';\
 	bash -c "diff -u <(echo -n) <(go fmt .)";\
	EXIT_CODE=$$?;\
	if [ "$$EXIT_CODE"  -ne 0 ]; then \
		echo '$@: Go files must be formatted with gofmt'; \
	fi && \
	exit $$EXIT_CODE

lint:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint
	golangci-lint run

fmt:
	$(GOFMT) .

test:
	$(GOFMT) ./...
	$(GOTEST) -race -covermode=atomic ./...

coverage: get test
	$(GOTEST) -race -coverprofile=coverage.txt -covermode=atomic .

BENCH_COUNT ?= 5
bench:
	$(GOGET) golang.org/x/perf/cmd/benchstat@latest
	$(GOTEST) -bench=. ./... -benchtime=5s -benchmem -count=6 -run=^#
