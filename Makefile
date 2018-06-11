GOOS ?= linux
ARCH ?= amd64
BUILD := $(shell git describe --always --dirty)
VERSION ?= ${BUILD}

.PHONY: all
all: test build

.PHONY: ci
ci: gofmt govet golint test

.PHONY: govet
govet:
	go vet $(shell go list ./... | grep -v vendor)

.PHONY: golint
golint:
	golint $(shell go list ./... | grep -v vendor)

.PHONY: gofmt
gofmt: # run in script cause gofmt will exit 0 even if files need formatting
	ci/gofmt.sh

.PHONY: test
test:
	@go test ./pkg/oneandone

.PHONY: build
build:
	@GOOS=${GOOS} GOARCH=${ARCH} CGO_ENABLED=0 go build \
	-ldflags "-X main.version=${VERSION} -X main.build=${BUILD}" \
	.