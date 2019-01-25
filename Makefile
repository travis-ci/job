GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
SOURCES := $(shell git ls-files '*.go')
TARGETS := \
	build/$(GOOS)/$(GOARCH)/travis-proc

.PHONY: all
all: build test
	
.PHONY: build
build: $(TARGETS)

build/$(GOOS)/$(GOARCH)/travis-proc: $(SOURCES)
	go build -o $@ ./cmd/travis-proc/...

.PHONY: clean
clean:
	rm -rf $(TARGETS)

.PHONY: test
test:
	go test -v ./...
