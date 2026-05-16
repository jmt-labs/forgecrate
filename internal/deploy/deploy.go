package deploy

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/jmt-labs/claude-setup/internal/compose"
	"github.com/jmt-labs/claude-setup/internal/config"
	"github.com/jmt-labs/claude-setup/internal/extensions"
)

func Run(sourceDir, destDir string, cfg config.Config) error {
	return RunWithClaude(sourceDir, destDir, cfg, "claude")
}

func RunWithClaude(sourceDir, destDir string, cfg config.Config, claudeBin string) error {
	req := compose.Request{
		SourceDir: sourceDir,
		DestDir:   destDir,
		Profile:   cfg.Profile,
		Flavors:   cfg.Flavors,
	}
	if err := compose.Run(req); err != nil {
		return fmt.Errorf("compose: %w", err)
	}

	if err := copyHooks(sourceDir, destDir); err != nil {
		return fmt.Errorf("hooks: %w", err)
	}

	if err := installExtensions(sourceDir, destDir, cfg, claudeBin); err != nil {
		return fmt.Errorf("extensions: %w", err)
	}

	if err := copySkills(sourceDir, destDir, cfg); err != nil {
		return fmt.Errorf("skills: %w", err)
	}

	cfgPath := filepath.Join(destDir, ".claude-setup.yaml")
	if err := config.Write(cfgPath, cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func installExtensions(sourceDir, destDir string, cfg config.Config, claudeBin string) error {
	paths := []string{
		filepath.Join(sourceDir, "base", "extensions.yaml"),
		filepath.Join(sourceDir, "profiles", cfg.Profile, "extensions.yaml"),
	}
	for _, flavor := range cfg.Flavors {
		paths = append(paths, filepath.Join(sourceDir, "flavors", flavor, "extensions.yaml"))
	}

	var layers []extensions.Extensions
	for _, path := range paths {
		ext, err := extensions.Load(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("load %s: %w", path, err)
		}
		layers = append(layers, ext)
	}

	merged := extensions.Merge(layers)
	return extensions.Installer{Claude: claudeBin, Dir: destDir}.Install(merged)
}

func copySkills(sourceDir, destDir string, cfg config.Config) error {
	dirs := []string{
		filepath.Join(sourceDir, "base", "skills"),
		filepath.Join(sourceDir, "profiles", cfg.Profile, "skills"),
	}
	for _, flavor := range cfg.Flavors {
		dirs = append(dirs, filepath.Join(sourceDir, "flavors", flavor, "skills"))
	}

	seen := map[string]bool{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read %s: %w", dir, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if seen[name] {
				continue
			}
			seen[name] = true
			src := filepath.Join(dir, name)
			dst := filepath.Join(destDir, ".claude", "skills", name)
			if err := copyDir(src, dst); err != nil {
				return fmt.Errorf("copy skill %s: %w", name, err)
			}
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
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return nil
}

func copyHooks(src, dst string) error {
	hooksDir := filepath.Join(src, "base", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		log.Printf("warn: hooks-Verzeichnis nicht gefunden (%s) — Hook-Dateien werden nicht deployt, aber settings.json referenziert sie", hooksDir)
		return nil
	}

	dstHooks := filepath.Join(dst, ".claude", "hooks")
	if err := os.MkdirAll(dstHooks, 0755); err != nil {
		return fmt.Errorf("mkdir hooks: %w", err)
	}

	return filepath.Walk(hooksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(hooksDir, path)
		return copyExecutable(path, filepath.Join(dstHooks, rel))
	})
}

func copyExecutable(src, dst string) (err error) {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return nil
}
