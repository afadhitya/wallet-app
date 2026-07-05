.PHONY: build run install test test-cover coverage-check lint fmt tidy verify-deps sqlc-gen docs clean

# Coverage notes:
# - internal/gen (sqlc-generated) is excluded via package filter
# - cmd/coverage-filter (tooling) is excluded via package filter
# - coverignore.txt documents OS/infrastructure error branches excluded from coverage gate
#   These are branches that require OS-level failure injection to trigger (e.g. config load,
#   MkdirAll, db.Open failures in CLI init). All business logic must still reach 100%.

build:
	go build -o bin/wallet ./cmd/wallet

run: build
	./bin/wallet

install:
	go install ./cmd/wallet

test:
	go test ./...

test-cover:
	go test -coverprofile=coverage.raw -covermode=atomic $$(go list ./... | grep -v '/internal/gen$$' | grep -v '/cmd/coverage-filter$$')
	go run cmd/coverage-filter/main.go coverignore.txt coverage.raw > coverage.out
	go tool cover -html=coverage.out -o coverage.html

coverage-check: test-cover
	@total=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ "$$total" != "100.0" ]; then \
		echo "ERROR: coverage is $$total%, expected 100.0%"; \
		echo "Run 'make test-cover' to see filtered coverage report."; \
		exit 1; \
	fi; \
	echo "Coverage: $$total% (infrastructure exclusions applied)"

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

docs:
	@mkdir -p docs/cli
	@go run cmd/wallet/main.go docs markdown
	@echo "Documentation generated in docs/cli/"

clean:
	rm -rf bin/ coverage.out coverage.html docs/cli/
