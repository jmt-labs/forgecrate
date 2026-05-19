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
	"strconv"
	"strings"
	"time"
)

var defaultRetryDelays = []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

// Client is an HTTP client for the GitHub API.
// RetryDelays controls the wait times between retries (default: 1s, 2s, 4s).
// Set to zero-duration slices in tests for instant retries.
type Client struct {
	baseURL     string
	httpClient  *http.Client
	token       string
	RetryDelays []time.Duration
}

func New(baseURL string) *Client {
	return &Client{
		baseURL:     baseURL,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		token:       os.Getenv("GITHUB_TOKEN"),
		RetryDelays: defaultRetryDelays,
	}
}

func Default() *Client {
	return &Client{
		baseURL:     "https://api.github.com",
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		token:       os.Getenv("GITHUB_TOKEN"),
		RetryDelays: defaultRetryDelays,
	}
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	delays := c.RetryDelays
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= len(delays); attempt++ {
		if attempt > 0 {
			time.Sleep(delays[attempt-1])
			newReq, cloneErr := http.NewRequestWithContext(req.Context(), req.Method, req.URL.String(), nil)
			if cloneErr != nil {
				return nil, fmt.Errorf("clone request: %w", cloneErr)
			}
			newReq.Header = req.Header.Clone()
			req = newReq
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			if attempt == len(delays) {
				return nil, err
			}
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			reset := resp.Header.Get("X-RateLimit-Reset")
			_ = resp.Body.Close()
			if reset != "" {
				ts, parseErr := strconv.ParseInt(reset, 10, 64)
				if parseErr == nil {
					resetTime := time.Unix(ts, 0).UTC()
					if attempt == len(delays) {
						return nil, fmt.Errorf("rate limit exceeded, resets at %s", resetTime.Format(time.RFC3339))
					}
					continue
				}
			}
			if attempt == len(delays) {
				return nil, fmt.Errorf("rate limit exceeded (HTTP 429)")
			}
			continue
		}

		if resp.StatusCode >= 500 {
			_ = resp.Body.Close()
			if attempt == len(delays) {
				return nil, fmt.Errorf("server error: HTTP %d", resp.StatusCode)
			}
			continue
		}

		return resp, nil
	}

	return resp, err
}

func (c *Client) Download(owner, repo, ref, destDir string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/tarball/%s", c.baseURL, owner, repo, ref)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := c.do(req)
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
