SRC := ./cmd
BIN := gomeet
FLAGS := "-s -w"

ifeq ($(OS),Windows_NT)
	BIN = gomeet.exe
	FLAGS = "-H windowsgui -s -w"
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