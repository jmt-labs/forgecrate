.PHONY: test test-e2e quality build clean

test:
	go test ./internal/... ./cmd/...

test-e2e:
	go test ./e2e/...

quality:
	go vet ./...
	go build ./...

build:
	go build -o claude-setup ./cmd/claude-setup/

clean:
	go clean -testcache
	rm -f claude-setup
