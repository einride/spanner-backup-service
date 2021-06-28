SHELL := /bin/bash

.PHONY: all
all: \
	go-test \
	go-mod-tidy \

include ./tools/golangci-lint/rules.mk
include ./tools/goreview/rules.mk
include ./tools/semantic-release/rules.mk


.PHONY: go-mod-tidy
go-mod-tidy:
	$(info [$@] tidying Go module files...)
	@find . -name go.mod -execdir go mod tidy \;

.PHONY: go-test
go-test:
	$(info [$@] running Go tests...)
	@go test -race -covermode=atomic ./...
