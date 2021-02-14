.PHONY: build run test

build:
	go build -o bin/main cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test -count=1 ./...