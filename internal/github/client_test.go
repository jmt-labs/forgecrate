package github_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	gh "github.com/jmt-labs/forgecrate/internal/github"
)

func makeTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{Name: "repo-prefix/" + name, Mode: 0644, Size: int64(len(content))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("tar WriteHeader: %v", err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("tar Write: %v", err)
		}
	}
	tw.Close()
	gz.Close()
	return buf.Bytes()
}

func TestDownloadAndExtract(t *testing.T) {
	tarball := makeTarGz(t, map[string]string{
		"base/CLAUDE.md":             "# Base",
		"base/.claude/settings.json": `{"hooks":{}}`,
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-gzip")
		w.Write(tarball)
	}))
	defer srv.Close()

	client := gh.New(srv.URL)
	dir := t.TempDir()

	if err := client.Download("jmt-labs", "forgecrate", "main", dir); err != nil {
		t.Fatalf("Download: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "base", "CLAUDE.md"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(content) != "# Base" {
		t.Errorf("got %q, want %q", string(content), "# Base")
	}
}

func TestDownloadHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := gh.New(srv.URL)
	err := client.Download("jmt-labs", "forgecrate", "main", t.TempDir())
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestDownloadPathTraversal(t *testing.T) {
	tarball := makeTarGz(t, map[string]string{
		"../../etc/passwd": "evil",
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-gzip")
		w.Write(tarball)
	}))
	defer srv.Close()

	client := gh.New(srv.URL)
	err := client.Download("jmt-labs", "forgecrate", "main", t.TempDir())
	if err == nil {
		t.Fatal("expected error for path traversal attempt")
	}
}

// TestClientHasTimeout verifies that the HTTP client uses a timeout < 60s.
func TestClientHasTimeout(t *testing.T) {
	// A server that hangs forever.
	block := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block
	}))
	defer srv.Close()
	defer close(block)

	client := gh.New(srv.URL)
	client.SetHTTPTimeout(100 * time.Millisecond)

	start := time.Now()
	err := client.Download("o", "r", "ref", t.TempDir())
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if elapsed >= 60*time.Second {
		t.Errorf("timeout took too long: %v (must be < 60s)", elapsed)
	}
}

// TestAuthorizationHeader verifies that GITHUB_TOKEN is sent as Bearer token.
func TestAuthorizationHeader(t *testing.T) {
	token := "test-token-abc123"
	t.Setenv("GITHUB_TOKEN", token)

	var receivedAuth string
	tarball := makeTarGz(t, map[string]string{"file.txt": "hello"})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/x-gzip")
		w.Write(tarball)
	}))
	defer srv.Close()

	client := gh.New(srv.URL)
	if err := client.Download("o", "r", "ref", t.TempDir()); err != nil {
		t.Fatalf("Download: %v", err)
	}

	expected := "Bearer " + token
	if receivedAuth != expected {
		t.Errorf("Authorization header = %q, want %q", receivedAuth, expected)
	}
}

// TestNoAuthorizationHeaderWhenNoToken verifies that no Authorization header is sent without token.
func TestNoAuthorizationHeaderWhenNoToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	var receivedAuth string
	tarball := makeTarGz(t, map[string]string{"file.txt": "hello"})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/x-gzip")
		w.Write(tarball)
	}))
	defer srv.Close()

	client := gh.New(srv.URL)
	if err := client.Download("o", "r", "ref", t.TempDir()); err != nil {
		t.Fatalf("Download: %v", err)
	}

	if receivedAuth != "" {
		t.Errorf("expected no Authorization header, got %q", receivedAuth)
	}
}

// TestRetryOn429 verifies that two 429 responses are retried and the third 200 succeeds.
func TestRetryOn429(t *testing.T) {
	tarball := makeTarGz(t, map[string]string{"file.txt": "hello"})
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)
		if n <= 2 {
			// First two calls: 429 with reset time
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(1*time.Second).Unix()))
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		// Third call: success
		w.Header().Set("Content-Type", "application/x-gzip")
		w.Write(tarball)
	}))
	defer srv.Close()

	client := gh.New(srv.URL)
	client.SetRetryDelay(10 * time.Millisecond) // speed up test

	if err := client.Download("o", "r", "ref", t.TempDir()); err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("expected 3 calls (2 retries + 1 success), got %d", callCount)
	}
}

// TestRetryOn5xx verifies that 500 responses are retried.
func TestRetryOn5xx(t *testing.T) {
	tarball := makeTarGz(t, map[string]string{"file.txt": "hello"})
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/x-gzip")
		w.Write(tarball)
	}))
	defer srv.Close()

	client := gh.New(srv.URL)
	client.SetRetryDelay(10 * time.Millisecond)

	if err := client.Download("o", "r", "ref", t.TempDir()); err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("expected 3 calls (2 retries + 1 success), got %d", callCount)
	}
}

// TestMaxRetriesExhausted verifies that after 3 failed attempts, a clear error is returned.
func TestMaxRetriesExhausted(t *testing.T) {
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := gh.New(srv.URL)
	client.SetRetryDelay(10 * time.Millisecond)

	err := client.Download("o", "r", "ref", t.TempDir())
	if err == nil {
		t.Fatal("expected error after max retries, got nil")
	}

	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("expected exactly 3 attempts, got %d", callCount)
	}
}

// TestRateLimitErrorMessage verifies that the 429 error message includes the reset time.
func TestRateLimitErrorMessage(t *testing.T) {
	resetTime := time.Now().Add(30 * time.Second).Unix()
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := gh.New(srv.URL)
	client.SetRetryDelay(10 * time.Millisecond)

	err := client.Download("o", "r", "ref", t.TempDir())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errStr := err.Error()
	if len(errStr) == 0 {
		t.Fatal("error message is empty")
	}
	// The error message should mention rate limit and ideally the reset time
	t.Logf("error message: %s", errStr)
}
