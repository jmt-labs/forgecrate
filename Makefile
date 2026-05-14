.PHONY: test test-e2e quality build release clean

test:
	go test ./internal/... ./cmd/...

test-e2e:
	go test ./e2e/...

quality:
	go vet ./...
	go build ./...

build:
	go build -o claude-setup ./cmd/claude-setup/

release:
	mkdir -p dist
	GOOS=linux   GOARCH=amd64 go build -o dist/claude-setup-linux-amd64       ./cmd/claude-setup/
	GOOS=linux   GOARCH=arm64 go build -o dist/claude-setup-linux-arm64       ./cmd/claude-setup/
	GOOS=windows GOARCH=amd64 go build -o dist/claude-setup-windows-amd64.exe ./cmd/claude-setup/
	GOOS=windows GOARCH=arm64 go build -o dist/claude-setup-windows-arm64.exe ./cmd/claude-setup/
	GOOS=darwin  GOARCH=arm64 go build -o dist/claude-setup-darwin-arm64      ./cmd/claude-setup/

clean:
	go clean -testcache
	rm -f claude-setup
	rm -rf dist/
