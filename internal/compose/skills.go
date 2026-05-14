package compose

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// MergeSkills kopiert Skill-Dateien aus allen srcDirs in destDir.
// Spätere srcDirs überschreiben frühere (Layer-Precedence).
// Dateien unter destDir/overrides/ werden nie überschrieben.
func MergeSkills(srcDirs []string, destDir string) error {
	for _, src := range srcDirs {
		if err := copyDir(src, destDir); err != nil {
			return err
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, rel)

		if isUnderOverrides(dst, dstPath) {
			return nil
		}

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}
		return copyFile(path, dstPath)
	})
}

func isUnderOverrides(destDir, path string) bool {
	overridesDir := filepath.Join(destDir, "overrides")
	rel, err := filepath.Rel(overridesDir, path)
	if err != nil {
		return false
	}
	return len(rel) > 0 && rel[0] != '.'
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return nil
}
