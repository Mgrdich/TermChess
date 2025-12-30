.PHONY: build test run clean

build:
	go build -o bin/termchess ./cmd/termchess

test:
	go test -v ./...

run:
	go run ./cmd/termchess

clean:
	rm -rf bin/
