SHELL := /bin/bash
GO ?= go
PKG := ./...
BINARY := go-sudoku
CLIBIN := sudoku-cli
OUT := bin

.PHONY: all fmt vet test cover build run tidy clean docker-build docker-run docker-push cli gui build-gui rebuild-gui

all: fmt vet test

fmt:
	$(GO) fmt $(PKG)

vet:
	$(GO) vet $(PKG)

# Runs tests with race detector and coverage
TEST_FLAGS ?= -race -coverprofile=coverage.out -covermode=atomic

test:
	$(GO) test $(PKG) $(TEST_FLAGS)

cover: test
	@$(GO) tool cover -func=coverage.out | tail -n 1
	@echo "HTML report: coverage.html"
	@$(GO) tool cover -html=coverage.out -o coverage.html

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

build:
	mkdir -p $(OUT)
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(OUT)/$(BINARY) ./cmd/server
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(OUT)/$(CLIBIN) ./cmd/cli

run:
	$(GO) run ./cmd/server

cli:
	$(GO) run ./cmd/cli

# GUI demo (optional; adds dependency fyne.io/fyne/v2 when built with -tags gui)
GUI_TAGS ?= gui

gui:
	$(GO) run -mod=mod -tags $(GUI_TAGS) ./cmd/gui

build-gui:
	mkdir -p $(OUT)
	$(GO) build -mod=mod -tags $(GUI_TAGS) -trimpath -ldflags "$(LDFLAGS)" -o $(OUT)/sudoku-gui ./cmd/gui

# Rebuild the GUI from scratch by removing the existing binary first
rebuild-gui:
	rm -f $(OUT)/sudoku-gui
	$(MAKE) build-gui

clean:
	rm -rf $(OUT) coverage.out coverage.html

TAGS?=
bench:
	$(GO) test -bench=. -benchmem $(PKG) -tags $(TAGS)

tidy:
	$(GO) mod tidy -v

IMAGE ?= ghcr.io/rumendamyanov/go-sudoku:latest

docker-build:
	docker build -t $(IMAGE) .

docker-run:
	docker run --rm -p 8080:8080 -e PORT=8080 $(IMAGE)

docker-push:
	docker push $(IMAGE)
