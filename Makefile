.PHONY: test test-e2e quality build release clean check-model-ids check-readme-coverage

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
	go build -o forgecrate ./cmd/forgecrate/

release:
	goreleaser release --clean

check-model-ids:
	@found=$$(grep -rn "claude-opus-4-7\|claude-sonnet-4-6\|claude-haiku-4-5" . \
		--include="*.md" --include="*.yaml" --include="*.json" --include="*.go" \
		--exclude-dir=".git" --exclude-dir="worktrees" --exclude-dir="dist" \
		--exclude-dir="docs" \
		| grep -v "base/models\.yaml" \
		| grep -v "base/\.claude/settings\.json" \
		| grep -v "\.claude/settings\.json" \
		| grep -v "CHANGELOG" \
		| grep -v "CLAUDE\.md" \
		| grep -v "README\.md" \
		| grep -v "_test\.go"); \
	if [ -n "$$found" ]; then \
		echo "ERROR: Model IDs found outside canonical source (base/models.yaml):"; \
		echo "$$found"; \
		exit 1; \
	fi
	@echo "OK: All model IDs are in canonical location (base/models.yaml)"

check-readme-coverage:
	@missing=0; \
	for flavor in $$(ls flavors/); do \
		if ! grep -q "$$flavor" README.md; then \
			echo "MISSING in README: flavors/$$flavor"; \
			missing=1; \
		fi; \
	done; \
	exit $$missing

clean:
	go clean -testcache
	rm -f forgecrate claude-setup
	rm -rf dist/
