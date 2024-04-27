-include $(PWD)/.env
-include $(PWD)/tools/Makefile
export

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor
	go mod verify

.PHONY: build
build:
	mkdir -p ./bin
	CGO_ENABLED=0 go build -o ./bin/api ./examples/api/main.go
	CGO_ENABLED=0 go build -o ./bin/worker ./examples/worker/main.go

.PHONY: unit-tests
unit-tests:
	@echo "--- Starting Unit Tests ---"
	go test -v -race -timeout 60s ./...
