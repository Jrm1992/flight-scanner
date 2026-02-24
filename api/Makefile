.PHONY: build run test lint fmt tidy clean docker-build

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test -race -cover ./cmd/... ./internal/...

lint:
	golangci-lint run ./cmd/... ./internal/...

fmt:
	gofmt -w ./cmd/ ./internal/

tidy:
	go mod tidy

clean:
	rm -rf bin/ coverage.out

docker-build:
	docker build -t flight-scanner:latest .
