.PHONY: build test lint coverage clean

build:
	go build -o edinet-cli .

test:
	go test -race ./...

lint:
	go vet ./...
	golangci-lint run ./...

coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -f edinet-cli coverage.out coverage.html
