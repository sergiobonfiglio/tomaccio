set dotenv-load := false

BIN := "tomaccio"

fmt:
    gofmt -w cmd internal

test:
    go test ./...

build:
    go build -o {{BIN}} ./cmd/tomaccio

release VERSION:
    go build -ldflags "-X github.com/sergiobonfiglio/tomaccio/internal/app.versionOverride={{VERSION}}" -o {{BIN}} ./cmd/tomaccio

check: fmt test build

run *ARGS:
    go run ./cmd/tomaccio {{ARGS}}
