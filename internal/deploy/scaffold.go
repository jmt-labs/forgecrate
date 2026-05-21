package deploy

import (
	"fmt"
	"os"
	"path/filepath"
)

func scaffoldMemoryBank(sourceDir, destDir string) error {
	srcDir := filepath.Join(sourceDir, "base", "memory-bank")
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil
	}

	dstDir := filepath.Join(destDir, "memory-bank")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("mkdir memory-bank: %w", err)
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("read memory-bank source: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		dst := filepath.Join(dstDir, entry.Name())
		if _, err := os.Stat(dst); err == nil {
			continue
		}
		if err := copyFile(filepath.Join(srcDir, entry.Name()), dst); err != nil {
			return fmt.Errorf("scaffold %s: %w", entry.Name(), err)
		}
	}
	return nil
}
