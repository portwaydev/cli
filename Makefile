APP_NAME := portway
BIN_DIR := bin

# Resolve version info from git
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Fallbacks when not in a git repo
ifeq ($(strip $(GIT_TAG)),)
  GIT_TAG := dev
endif
ifeq ($(strip $(GIT_COMMIT)),)
  GIT_COMMIT := none
endif

LDFLAGS := -s -w \
  -X 'cli/pkg/buildinfo.Version=$(GIT_TAG)' \
  -X 'cli/pkg/buildinfo.Commit=$(GIT_COMMIT)' \
  -X 'cli/pkg/buildinfo.Date=$(BUILD_DATE)'

.PHONY: build
build: $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME)

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

.PHONY: clean
clean:
	rm -rf bin

.PHONY: version
version: build
	./$(BIN_DIR)/$(APP_NAME) version || true


