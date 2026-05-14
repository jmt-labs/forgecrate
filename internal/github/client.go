package github

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Client struct {
	baseURL string
}

func New(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

func Default() *Client {
	return &Client{baseURL: "https://api.github.com"}
}

func (c *Client) Download(owner, repo, ref, destDir string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/tarball/%s", c.baseURL, owner, repo, ref)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return extractTarGz(resp.Body, destDir)
}

func extractTarGz(r io.Reader, destDir string) (err error) {
	gz, gzErr := gzip.NewReader(r)
	if gzErr != nil {
		return fmt.Errorf("gzip: %w", gzErr)
	}
	defer func() {
		if cerr := gz.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("gzip close: %w", cerr)
		}
	}()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("tar: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		// Strip leading path component (GitHub adds "owner-repo-sha/" prefix)
		parts := strings.SplitN(hdr.Name, "/", 2)
		if len(parts) < 2 {
			continue
		}
		rel := parts[1]
		if rel == "" {
			continue
		}

		dst := filepath.Join(destDir, filepath.FromSlash(rel))
		if !strings.HasPrefix(dst, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("tar: illegal path %q", hdr.Name)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return fmt.Errorf("mkdir: %w", err)
		}
		f, err := os.Create(dst)
		if err != nil {
			return fmt.Errorf("create %s: %w", dst, err)
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	return nil
}
