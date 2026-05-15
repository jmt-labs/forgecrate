package deploy

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmt-labs/claude-setup/internal/config"
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
		return writeAndRecord(dstPath, rel, newContent, hashNew, cfg)
	}

	// Migration: kein stored hash → einfach überschreiben
	if !hasStored {
		return writeAndRecord(dstPath, rel, newContent, hashNew, cfg)
	}

	diskUnchanged := hashDisk == hashStored
	newSameAsDisk := hashNew == hashDisk

	if diskUnchanged {
		if newSameAsDisk {
			// Fall 1: unverändert, neue Version identisch → nichts tun
			return nil
		}
		// Fall 2: unverändert, neue Version verschieden → überschreiben
		return writeAndRecord(dstPath, rel, newContent, hashNew, cfg)
	}

	// Nutzer hat Datei geändert
	if newSameAsDisk {
		// Fall 3: Nutzer hat gleich geändert wie neue Version → nichts tun
		return nil
	}

	// Fall 4: echter Konflikt → prompt
	diskData, _ := os.ReadFile(dstPath)
	fmt.Fprintf(w, "\nKONFLIKT: %s\n", rel)
	fmt.Fprintf(w, "  Deine Version: %s\n", firstLine(diskData))
	fmt.Fprintf(w, "  Neue Version:  %s\n", firstLine(newContent))
	fmt.Fprintf(w, "  [ü]berschreiben / [b]ehalten (Standard: behalten): ")

	scanner := bufio.NewScanner(r)
	scanner.Scan()
	answer := strings.TrimSpace(scanner.Text())

	if answer == "ü" || answer == "u" {
		return writeAndRecord(dstPath, rel, newContent, hashNew, cfg)
	}
	// behalten: Hash der Nutzer-Version speichern
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
