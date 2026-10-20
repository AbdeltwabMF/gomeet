SRC := ./cmd
BIN := gomeet
FLAGS := "-s -w"

ifeq ($(OS),Windows_NT)
	BIN = gomeet.exe
endif

all:
		@go version
		@echo source: $(SRC)
		@echo binary: $(BIN)
		@echo flags:  $(FLAGS)

build:
		go build -o ./bin/$(BIN) -v -ldflags $(FLAGS) $(SRC)

run:
		go run $(SRC)

test:
		go test ./...

cover:
		go test -coverprofile=coverage.out ./...
		go tool cover -html=coverage.out

vet:
		go vet ./...

lint:
		golangci-lint run --timeout=30m

fmt:
		gofmt -w .

fmt-check:
		@test -z "$$(gofmt -l .)" || (echo "The following files are not formatted:"; gofmt -l .; exit 1)

check: fmt-check vet test lint
