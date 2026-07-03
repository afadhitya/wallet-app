.PHONY: build run install test test-cover coverage-check lint fmt tidy verify-deps sqlc-gen clean

build:
	go build -o bin/wallet ./cmd/wallet

run: build
	./bin/wallet

install:
	go install ./cmd/wallet

test:
	go test ./...

test-cover:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

coverage-check: test-cover
	@total=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ "$$total" != "100.0" ]; then \
		echo "ERROR: coverage is $$total%, expected 100.0%"; \
		go tool cover -html=coverage.out -o coverage.html; \
		exit 1; \
	fi; \
	echo "Coverage: $$total%"

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

verify-deps:
	go mod verify

sqlc-gen:
	sqlc generate

clean:
	rm -rf bin/ coverage.out coverage.html
