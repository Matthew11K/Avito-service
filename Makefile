COVERAGE_FILE ?= coverage.out

## test: run all tests
.PHONY: test
test:
	@go test -coverpkg='avito/...' --race -count=1 -coverprofile='$(COVERAGE_FILE)' ./...
	@go tool cover -func='$(COVERAGE_FILE)' | grep ^total | tr -s '\t'

.PHONY: lint
lint: lint-golang lint-proto

.PHONY: lint-golang
lint-golang:
	@if ! command -v 'golangci-lint' &> /dev/null; then \
  		echo "Please install golangci-lint!"; exit 1; \
  	fi;
	@golangci-lint -v run --fix ./...

.PHONY: lint-proto
lint-proto:
	@if ! command -v 'easyp' &> /dev/null; then \
  		echo "Please install easyp!"; exit 1; \
	fi;
	@easyp lint

.PHONY: gen-grpc
gen-grpc:
	@protoc --go_out=. --go-grpc_out=. api/proto/v1/pvz.proto

.PHONY: gen-api
gen-api:
	@mkdir -p internal/interfaces/http/dto
	@oapi-codegen -package dto -generate types -o internal/interfaces/http/dto/models.gen.go api/openapi/v1/swagger.yaml
	@echo "Generated API types from OpenAPI specification"

.PHONY: test-grpc
test-grpc:
	@go run cmd/grpctest/main.go

.PHONY: clean
clean:
	@rm -rf./bin