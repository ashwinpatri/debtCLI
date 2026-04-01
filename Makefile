.PHONY: build test lint install clean

build:
	go build -ldflags="-s -w" -o bin/debt .

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

install:
	go install .

clean:
	rm -rf bin/
