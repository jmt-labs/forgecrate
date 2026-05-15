package deploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashBytes(t *testing.T) {
	h := hashBytes([]byte("hello"))
	if h != "sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Errorf("unexpected hash: %s", h)
	}
}

func TestHashBytesEmpty(t *testing.T) {
	h := hashBytes([]byte{})
	if h != "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Errorf("unexpected hash: %s", h)
	}
}

func TestHashFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.txt")
	os.WriteFile(f, []byte("hello"), 0644)

	h, err := hashFile(f)
	if err != nil {
		t.Fatalf("hashFile: %v", err)
	}
	if h != hashBytes([]byte("hello")) {
		t.Errorf("hashFile != hashBytes for same content")
	}
}

func TestHashFileMissing(t *testing.T) {
	_, err := hashFile("/no/such/file")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
