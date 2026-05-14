package deploy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/markus/claude-setup/internal/compose"
	"github.com/markus/claude-setup/internal/config"
	"github.com/markus/claude-setup/internal/extensions"
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

	if err := installExtensions(sourceDir, cfg, claudeBin); err != nil {
		return fmt.Errorf("extensions: %w", err)
	}

	cfgPath := filepath.Join(destDir, ".claude-setup.yaml")
	if err := config.Write(cfgPath, cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func installExtensions(sourceDir string, cfg config.Config, claudeBin string) error {
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
	return extensions.Installer{Claude: claudeBin}.Install(merged)
}

func copyHooks(src, dst string) error {
	hooksDir := filepath.Join(src, "base", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
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
