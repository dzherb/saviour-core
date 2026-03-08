.PHONY: run-core
run-core:
	@go tool air -c .air.toml -- run --config ./configs/config.dev.json

.PHONY: test-fast
test-fast:
	@echo "skipping integration tests..."
	@go test -short -count=1 ./...

.PHONY: test
test:
	@go test -count=1 ./...

.PHONY: lint
lint:
	@golangci-lint run

.PHONY: lint-openapi
lint-openapi:
	@npx @redocly/cli@latest --config=./redocly.yaml lint ./api/openapi.yaml

.PHONY: fmt
fmt:
	@golangci-lint fmt

.PHONY: gen-server
gen-server:
	@go generate ./internal/transport/rest/api

.PHONY: gen-queries
gen-queries:
	@go generate ./internal/repository/...

.PHONY: gen-mocks
gen-mocks:
	@go generate ./internal/testkit/testmock

.PHONY: gen-all
gen-all:
	@go generate ./...
