.PHONY: run dev install-air test fmt vet

AIR_VERSION ?= v1.61.7
AIR_BIN := $(shell go env GOPATH)/bin/air

run:
	go run ./cmd/api

dev:
	@if command -v air >/dev/null 2>&1; then \
		air -c .air.toml; \
	elif [ -x "$(AIR_BIN)" ]; then \
		"$(AIR_BIN)" -c .air.toml; \
	else \
		go run github.com/air-verse/air@$(AIR_VERSION) -c .air.toml; \
	fi

install-air:
	go install github.com/air-verse/air@$(AIR_VERSION)

test:
	go test ./...

fmt:
	gofmt -w $$(find . -name '*.go' -type f)

vet:
	go vet ./...
