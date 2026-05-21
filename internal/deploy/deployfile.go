package deploy

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmt-labs/forgecrate/internal/config"
)

func deployFile(dstPath, rel string, newContent []byte, cfg *config.Config, w io.Writer, r io.Reader) error {
	if cfg.DeployedFiles == nil {
		cfg.DeployedFiles = map[string]string{}
	}

	hashNew := hashBytes(newContent)

	hashDisk := ""
	if diskData, err := os.ReadFile(dstPath); err == nil {
		hashDisk = hashBytes(diskData)
	}

	hashStored, hasStored := cfg.DeployedFiles[rel]

	// Datei existiert nicht → einfach schreiben
	if hashDisk == "" {
		_, _ = fmt.Fprintf(w, "✅ %s\n", rel)
		return writeAndRecord(dstPath, rel, newContent, hashNew, cfg)
	}

	// Migration: kein stored hash → einfach überschreiben
	if !hasStored {
		_, _ = fmt.Fprintf(w, "✅ %s\n", rel)
		return writeAndRecord(dstPath, rel, newContent, hashNew, cfg)
	}

	diskUnchanged := hashDisk == hashStored
	newSameAsDisk := hashNew == hashDisk

	if diskUnchanged {
		if newSameAsDisk {
			// Fall 1: unverändert, neue Version identisch → nichts tun
			_, _ = fmt.Fprintf(w, "🔵 %s\n", rel)
			return nil
		}
		// Fall 2: unverändert, neue Version verschieden → überschreiben
		_, _ = fmt.Fprintf(w, "✅ %s\n", rel)
		return writeAndRecord(dstPath, rel, newContent, hashNew, cfg)
	}

	// Nutzer hat Datei geändert
	if newSameAsDisk {
		// Fall 3: Nutzer hat gleich geändert wie neue Version → nichts tun, aber Hash aktualisieren
		_, _ = fmt.Fprintf(w, "🔵 %s\n", rel)
		cfg.DeployedFiles[rel] = hashDisk
		return nil
	}

	// Fall 4: echter Konflikt → prompt
	diskData, _ := os.ReadFile(dstPath)
	_, _ = fmt.Fprintf(w, "\nKONFLIKT: %s\n", rel)
	_, _ = fmt.Fprintf(w, "  Deine Version: %s\n", firstLine(diskData))
	_, _ = fmt.Fprintf(w, "  Neue Version:  %s\n", firstLine(newContent))
	_, _ = fmt.Fprintf(w, "  [o]verwrite / [k]eep (default: keep): ")

	scanner := bufio.NewScanner(r)
	scanner.Scan()
	answer := strings.TrimSpace(scanner.Text())

	if answer == "o" || answer == "ü" || answer == "u" {
		_, _ = fmt.Fprintf(w, "✅ %s  (conflict → replaced)\n", rel)
		return writeAndRecord(dstPath, rel, newContent, hashNew, cfg)
	}
	// behalten: Hash der Nutzer-Version speichern
	_, _ = fmt.Fprintf(w, "🔵 %s  (conflict → kept)\n", rel)
	cfg.DeployedFiles[rel] = hashDisk
	return nil
}

func writeAndRecord(dstPath, rel string, content []byte, hash string, cfg *config.Config) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(dstPath, content, 0644); err != nil {
		return err
	}
	cfg.DeployedFiles[rel] = hash
	return nil
}

func firstLine(data []byte) string {
	s := strings.TrimSpace(string(data))
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		s = s[:idx]
	}
	if len(s) > 80 {
		s = s[:80] + "…"
	}
	return s
}
