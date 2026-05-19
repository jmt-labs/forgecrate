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

const (
	defaultTimeout    = 30 * time.Second
	defaultRetryDelay = 1 * time.Second
	maxRetries        = 3
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	retryDelay time.Duration
}

func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
		retryDelay: defaultRetryDelay,
	}
}

func Default() *Client {
	return New("https://api.github.com")
}

// SetHTTPTimeout overrides the HTTP client timeout. Useful for testing.
func (c *Client) SetHTTPTimeout(d time.Duration) {
	c.httpClient.Timeout = d
}

// SetRetryDelay overrides the base delay between retries. Useful for testing.
func (c *Client) SetRetryDelay(d time.Duration) {
	c.retryDelay = d
}

func (c *Client) Download(owner, repo, ref, destDir string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/tarball/%s", c.baseURL, owner, repo, ref)

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := c.retryDelay * time.Duration(1<<uint(attempt-1))
			time.Sleep(delay)
		}

		resp, err := c.do(url)
		if err != nil {
			return fmt.Errorf("http get: %w", err)
		}

		switch {
		case resp.StatusCode == http.StatusOK:
			defer resp.Body.Close()
			return extractTarGz(resp.Body, destDir)

		case resp.StatusCode == http.StatusTooManyRequests:
			reset := resp.Header.Get("X-RateLimit-Reset")
			resp.Body.Close()
			if reset != "" {
				ts, parseErr := strconv.ParseInt(reset, 10, 64)
				if parseErr == nil {
					resetTime := time.Unix(ts, 0)
					lastErr = fmt.Errorf("rate limit exceeded (429): resets at %s", resetTime.UTC().Format(time.RFC3339))
				} else {
					lastErr = fmt.Errorf("rate limit exceeded (429): X-RateLimit-Reset=%s", reset)
				}
			} else {
				lastErr = fmt.Errorf("rate limit exceeded (429)")
			}

		case resp.StatusCode >= 500:
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %s", resp.Status)

		default:
			resp.Body.Close()
			return fmt.Errorf("unexpected status: %s", resp.Status)
		}
	}

	return fmt.Errorf("after %d attempts: %w", maxRetries, lastErr)
}

func (c *Client) do(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return c.httpClient.Do(req)
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
