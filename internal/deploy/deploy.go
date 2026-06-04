package deploy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jmt-labs/forgecrate/internal/compose"
	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/jmt-labs/forgecrate/internal/extensions"
)

func Run(sourceDir, destDir string, cfg config.Config) error {
	claudeBin := os.Getenv("CLAUDE_BIN")
	if claudeBin == "" {
		claudeBin = "claude"
	}
	return RunWithClaude(sourceDir, destDir, cfg, claudeBin, os.Stdout, os.Stdin)
}

func codegraphBin() string {
	if bin := os.Getenv("CODEGRAPH_BIN"); bin != "" {
		return bin
	}
	return "codegraph"
}

func RunWithClaude(sourceDir, destDir string, cfg config.Config, claudeBin string, out io.Writer, in io.Reader) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := validateSelection(sourceDir, cfg); err != nil {
		return err
	}
	req := compose.Request{
		SourceDir:      sourceDir,
		DestDir:        destDir,
		Profile:        cfg.Profile,
		Flavors:        cfg.Flavors,
		PermissionMode: cfg.PermissionMode,
		SkipSettings:   true,
	}

	// Settings: Inhalt berechnen, dann konflikt-sicher schreiben
	settingsContent, err := compose.ComposeSettings(req)
	if err != nil {
		return fmt.Errorf("compose settings: %w", err)
	}
	settingsPath := filepath.Join(destDir, ".claude", "settings.json")
	if err := deployFile(settingsPath, ".claude/settings.json", settingsContent, &cfg, out, in); err != nil {
		return fmt.Errorf("settings: %w", err)
	}

	if err := composeWithLog(req, out); err != nil {
		return fmt.Errorf("compose: %w", err)
	}

	if err := copyHooks(sourceDir, destDir, &cfg, out, in); err != nil {
		return fmt.Errorf("hooks: %w", err)
	}

	if err := installExtensions(sourceDir, destDir, cfg, claudeBin, out); err != nil {
		return fmt.Errorf("extensions: %w", err)
	}

	if err := scaffoldMemoryBank(sourceDir, destDir); err != nil {
		return fmt.Errorf("memory-bank scaffold: %w", err)
	}

	if err := copySkills(sourceDir, destDir, cfg, out); err != nil {
		return fmt.Errorf("skills: %w", err)
	}

	if err := appendFlavorGitignores(sourceDir, destDir, cfg); err != nil {
		return fmt.Errorf("gitignore: %w", err)
	}

	cfgPath := filepath.Join(destDir, ".forgecrate.yaml")
	if err := config.Write(cfgPath, cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func composeWithLog(req compose.Request, out io.Writer) error {
	for _, f := range []string{"CLAUDE.md", "AGENTS.md"} {
		destPath := filepath.Join(req.DestDir, f)
		before := fileHash(destPath)
		if err := compose.RunSingle(req, f); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("%s: %w", f, err)
		}
		after := fileHash(destPath)
		if before == after {
			_, _ = fmt.Fprintf(out, "🔵 %s\n", f)
		} else {
			_, _ = fmt.Fprintf(out, "✅ %s\n", f)
		}
	}
	return compose.RunCommands(req)
}

func fileHash(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return hashBytes(data)
}

func installExtensions(sourceDir, destDir string, cfg config.Config, claudeBin string, out io.Writer) error {
	if err := extensions.CheckNodeVersion(18); err != nil {
		return fmt.Errorf("node-check: %w", err)
	}


	profileCfg := compose.LoadProfileConfig(sourceDir, cfg.Profile)

	paths := []string{
		filepath.Join(sourceDir, "base", "extensions.yaml"),
	}
	for _, parent := range profileCfg.Extends {
		paths = append(paths, filepath.Join(sourceDir, "profiles", parent, "extensions.yaml"))
	}
	paths = append(paths, filepath.Join(sourceDir, "profiles", cfg.Profile, "extensions.yaml"))
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
	if err := extensions.WriteMCPJson(destDir, merged); err != nil {
		return fmt.Errorf("write .mcp.json: %w", err)
	}
	installer := extensions.Installer{Claude: claudeBin, Dir: destDir, Out: out}
	if err := installer.Install(merged); err != nil {
		return err
	}

	if slices.Contains(cfg.Flavors, "codegraph") {
		if err := extensions.InitCodegraph(destDir, codegraphBin(), out); err != nil {
			return fmt.Errorf("codegraph init: %w", err)
		}
	}

	return nil
}

func copySkills(sourceDir, destDir string, cfg config.Config, out io.Writer) error {
	profileCfg := compose.LoadProfileConfig(sourceDir, cfg.Profile)

	dirs := []string{
		filepath.Join(sourceDir, "base", "skills"),
	}
	for _, parent := range profileCfg.Extends {
		dirs = append(dirs, filepath.Join(sourceDir, "profiles", parent, "skills"))
	}
	dirs = append(dirs, filepath.Join(sourceDir, "profiles", cfg.Profile, "skills"))
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
			_, _ = fmt.Fprintf(out, "✅ skill:%s\n", name)
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
	defer func() { _ = in.Close() }()

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

func copyHooks(src, dst string, cfg *config.Config, out io.Writer, in io.Reader) error {
	hooksDir := filepath.Join(src, "base", "hooks")
	if _, err := os.Stat(hooksDir); err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintf(out, "🔵 hooks: kein Verzeichnis vorhanden, wird übersprungen\n")
			return nil
		}
		return fmt.Errorf("hooks-Verzeichnis prüfen (%s): %w", hooksDir, err)
	}

	dstHooks := filepath.Join(dst, ".claude", "hooks")
	if err := os.MkdirAll(dstHooks, 0755); err != nil {
		return fmt.Errorf("mkdir hooks: %w", err)
	}

	if err := walkHooksDir(hooksDir, dstHooks, cfg, out, in); err != nil {
		return err
	}

	for _, flavor := range cfg.Flavors {
		flavorHooksDir := filepath.Join(src, "flavors", flavor, "hooks")
		if _, err := os.Stat(flavorHooksDir); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return fmt.Errorf("flavor-hooks-Verzeichnis prüfen (%s): %w", flavorHooksDir, err)
		}
		if err := walkHooksDir(flavorHooksDir, dstHooks, cfg, out, in); err != nil {
			return err
		}
	}
	return nil
}

func appendFlavorGitignores(sourceDir, destDir string, cfg config.Config) error {
	gitignorePath := filepath.Join(destDir, ".gitignore")
	existing, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read .gitignore: %w", err)
	}
	current := string(existing)

	f, err := os.OpenFile(gitignorePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open .gitignore: %w", err)
	}
	defer func() { _ = f.Close() }()

	for _, flavor := range cfg.Flavors {
		beginMarker := "# forgecrate:" + flavor + " BEGIN"
		if strings.Contains(current, beginMarker) {
			continue
		}
		srcPath := filepath.Join(sourceDir, "flavors", flavor, "gitignore.txt")
		data, err := os.ReadFile(srcPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read %s: %w", srcPath, err)
		}
		content := strings.TrimRight(string(data), "\n")
		if content == "" {
			continue
		}
		block := beginMarker + "\n" + content + "\n# forgecrate:" + flavor + " END\n"
		if _, err := fmt.Fprint(f, block); err != nil {
			return fmt.Errorf("write .gitignore: %w", err)
		}
		current += block
	}
	return nil
}

func walkHooksDir(hooksDir, dstHooks string, cfg *config.Config, out io.Writer, in io.Reader) error {
	return filepath.Walk(hooksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(hooksDir, path)
		dstPath := filepath.Join(dstHooks, rel)

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read hook %s: %w", rel, err)
		}

		relKey := filepath.Join(".claude", "hooks", rel)
		if err := deployFile(dstPath, relKey, content, cfg, out, in); err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		return os.Chmod(dstPath, 0755)
	})
}
