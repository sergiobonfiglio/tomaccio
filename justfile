set dotenv-load := false

BIN := "tomaccio"

fmt:
    gofmt -w cmd internal

test:
    go test ./...

build:
    go build -o {{BIN}} ./cmd/tomaccio

check: fmt test build

run *ARGS:
    go run ./cmd/tomaccio {{ARGS}}
