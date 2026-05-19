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

// noDelay returns a Client with zero retry delays for fast tests.
func noDelay(baseURL string) *gh.Client {
	c := gh.New(baseURL)
	c.RetryDelays = []time.Duration{0, 0, 0}
	return c
}

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
		"base/CLAUDE.md":            "# Base",
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

func TestDownloadSendsAuthHeader(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-secret-token")

	tarball := makeTarGz(t, map[string]string{"file.txt": "hello"})

	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/x-gzip")
		w.Write(tarball)
	}))
	defer srv.Close()

	client := gh.New(srv.URL)
	if err := client.Download("owner", "repo", "main", t.TempDir()); err != nil {
		t.Fatalf("Download: %v", err)
	}

	want := "Bearer test-secret-token"
	if gotAuth != want {
		t.Errorf("Authorization header = %q, want %q", gotAuth, want)
	}
}

func TestDownloadNoAuthHeaderWhenTokenUnset(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	tarball := makeTarGz(t, map[string]string{"file.txt": "hello"})

	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/x-gzip")
		w.Write(tarball)
	}))
	defer srv.Close()

	client := gh.New(srv.URL)
	if err := client.Download("owner", "repo", "main", t.TempDir()); err != nil {
		t.Fatalf("Download: %v", err)
	}

	if gotAuth != "" {
		t.Errorf("expected no Authorization header, got %q", gotAuth)
	}
}

func TestDownloadRetryOn429ThenSuccess(t *testing.T) {
	tarball := makeTarGz(t, map[string]string{"file.txt": "retried"})

	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n == 1 {
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", 0))
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/x-gzip")
		w.Write(tarball)
	}))
	defer srv.Close()

	client := noDelay(srv.URL)
	dir := t.TempDir()
	if err := client.Download("owner", "repo", "main", dir); err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if attempts.Load() != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts.Load())
	}
}

func TestDownloadRetryOn5xxThenSuccess(t *testing.T) {
	tarball := makeTarGz(t, map[string]string{"file.txt": "retried-5xx"})

	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/x-gzip")
		w.Write(tarball)
	}))
	defer srv.Close()

	client := noDelay(srv.URL)
	dir := t.TempDir()
	if err := client.Download("owner", "repo", "main", dir); err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestDownloadRateLimitExhausted(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.Header().Set("X-RateLimit-Reset", "1999999999")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := noDelay(srv.URL)
	err := client.Download("owner", "repo", "main", t.TempDir())
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if attempts.Load() != 4 {
		t.Errorf("expected 4 attempts (1 + 3 retries), got %d", attempts.Load())
	}
}
