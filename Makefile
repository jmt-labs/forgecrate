.PHONY: test test-e2e quality build release clean

test:
	go test ./internal/... ./cmd/...

test-e2e:
	@if [ -z "$$CLAUDE_BIN" ]; then \
		FAKE=$$(mktemp); \
		printf '#!/bin/sh\nexit 0\n' > "$$FAKE"; \
		chmod +x "$$FAKE"; \
		CLAUDE_BIN="$$FAKE" go test ./e2e/...; \
		EXIT=$$?; rm -f "$$FAKE"; exit $$EXIT; \
	else \
		go test ./e2e/...; \
	fi

quality:
	go vet ./...
	go build ./...

build:
	go build -o claude-setup ./cmd/claude-setup/

release:
	goreleaser release --clean

clean:
	go clean -testcache
	rm -f claude-setup
	rm -rf dist/
