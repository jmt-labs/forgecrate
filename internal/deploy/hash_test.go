package deploy

import (
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

