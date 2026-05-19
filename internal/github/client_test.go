package github_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

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
