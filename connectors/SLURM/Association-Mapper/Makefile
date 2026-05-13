.PHONY: build test run tidy clean

BIN := bin/association-mapper

build:
	go build -o $(BIN) .

run: build
	./$(BIN)

test:
	go test ./...

tidy:
	go mod tidy

clean:
	rm -rf bin
