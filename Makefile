# Copyright 2024 Thales
.PHONY: all lint build coverage dev gen

# Project name
PROJECT_NAME := freyja
GO_MODULE_NAME := "github.com/ThalesGroup/$(PROJECT_NAME)"
CMD_PATH := cmd/$(PROJECT_NAME)/main.go
DIST_PATH := dist
BIN_PATH := $(DIST_PATH)/$(PROJECT_NAME)
COV_PATH := $(DIST_PATH)/coverage.out

# Useful variables for build metadata
VERSION ?= $(shell git describe --tags --always)
COMMIT_LONG ?= $(shell git rev-parse HEAD)
COMMIT_SHORT ?= $(shell git rev-parse --short=8 HEAD)
COMMIT_TIMESTAMP := $(shell git show -s --format=%cI HEAD)
GO_VERSION ?= $(shell go version)
BUILD_PLATFORM  ?= $(shell uname -m)
BUILD_DATE ?= $(shell date -u --iso-8601=seconds)
LDFLAGS = "-X '$(GO_MODULE_NAME)/pkg/version.RawGitDescribe=$(VERSION)' -X '$(GO_MODULE_NAME)/pkg/version.GitCommitIdLong=$(COMMIT_LONG)' -X '$(GO_MODULE_NAME)/pkg/version.GitCommitIdShort=$(COMMIT_SHORT)' -X '$(GO_MODULE_NAME)/pkg/version.GoVersion=$(GO_VERSION)' -X '$(GO_MODULE_NAME)/pkg/version.BuildPlatform=$(BUILD_PLATFORM)' -X '$(GO_MODULE_NAME)/pkg/version.BuildDate=$(BUILD_DATE)' -X '$(GO_MODULE_NAME)/pkg/version.GitCommitTimestamp=$(COMMIT_TIMESTAMP)'"
GO_LDFLAGS = -ldflags=$(LDFLAGS)
BINARY_NAME = $(PROJECT_NAME)

build:
		@go version
		@go build $(GO_LDFLAGS) -o $(BIN_PATH) $(CMD_PATH)
build-debug:
		@go version
		@go build -gcflags="all=-N -l" -o $(BIN_PATH) $(CMD_PATH)
		$(info use cmd : dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec k8s-kms-plugin)
		$(info will listen to port 2345)
coverage:
		mkdir -p $(DIST_PATH)
		go test -race -v -coverprofile $(COV_PATH) ./test/...
		go tool cover -html=$(COV_PATH) -o $(DIST_PATH)/coverage.html