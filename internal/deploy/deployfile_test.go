package deploy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
)

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// Fall 1: disk == stored, new == disk → nichts tun
func TestDeployFileNoChange(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "file.txt")
	writeTestFile(t, dst, "original")

	cfg := &config.Config{
		DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
	}

	err := deployFile(dst, "file.txt", []byte("original"), cfg, &strings.Builder{}, strings.NewReader(""))
	if err != nil {
		t.Fatalf("deployFile: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "original" {
		t.Errorf("file should be unchanged: %q", data)
	}
}

// Fall 2: disk == stored, new != disk → einfach überschreiben (kein Prompt)
func TestDeployFileCleanUpdate(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "file.txt")
	writeTestFile(t, dst, "original")

	cfg := &config.Config{
		DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
	}

	out := &strings.Builder{}
	err := deployFile(dst, "file.txt", []byte("updated"), cfg, out, strings.NewReader(""))
	if err != nil {
		t.Fatalf("deployFile: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "updated" {
		t.Errorf("expected updated content, got %q", data)
	}
	if strings.Contains(out.String(), "KONFLIKT") {
		t.Error("clean update should not show conflict")
	}
	if cfg.DeployedFiles["file.txt"] != hashBytes([]byte("updated")) {
		t.Error("hash not updated in cfg")
	}
}

// Fall 3: disk != stored, new == disk → Nutzer hat geändert, neue Version identisch → nichts tun
func TestDeployFileUserChangedSameAsNew(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "file.txt")
	writeTestFile(t, dst, "user-modified")

	cfg := &config.Config{
		DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
	}

	err := deployFile(dst, "file.txt", []byte("user-modified"), cfg, &strings.Builder{}, strings.NewReader(""))
	if err != nil {
		t.Fatalf("deployFile: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "user-modified" {
		t.Errorf("file should be preserved: %q", data)
	}
	// Hash muss auf den aktuellen Disk-Hash aktualisiert werden
	if cfg.DeployedFiles["file.txt"] != hashBytes([]byte("user-modified")) {
		t.Errorf("hash should be updated to current disk hash, got: %q", cfg.DeployedFiles["file.txt"])
	}
}

// Fall 4a: Konflikt → Nutzer wählt behalten
func TestDeployFileConflictKeep(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "file.txt")
	writeTestFile(t, dst, "user-modified")

	cfg := &config.Config{
		DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
	}

	out := &strings.Builder{}
	err := deployFile(dst, "file.txt", []byte("remote-new"), cfg, out, strings.NewReader("b\n"))
	if err != nil {
		t.Fatalf("deployFile: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "user-modified" {
		t.Errorf("file should be kept: %q", data)
	}
	if !strings.Contains(out.String(), "KONFLIKT") {
		t.Error("conflict should be reported")
	}
}

// Fall 4b: Konflikt → Nutzer wählt überschreiben mit "ü"
func TestDeployFileConflictOverwrite(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "file.txt")
	writeTestFile(t, dst, "user-modified")

	cfg := &config.Config{
		DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
	}

	out := &strings.Builder{}
	err := deployFile(dst, "file.txt", []byte("remote-new"), cfg, out, strings.NewReader("ü\n"))
	if err != nil {
		t.Fatalf("deployFile: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "remote-new" {
		t.Errorf("file should be overwritten: %q", data)
	}
	if cfg.DeployedFiles["file.txt"] != hashBytes([]byte("remote-new")) {
		t.Error("hash not updated after overwrite")
	}
}

// Fall 4b-neu: Konflikt → Nutzer wählt überschreiben mit "o" (primäres ASCII-Mnemonic)
func TestDeployFileConflictOverwriteWithO(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "file.txt")
	writeTestFile(t, dst, "user-modified")

	cfg := &config.Config{
		DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
	}

	out := &strings.Builder{}
	err := deployFile(dst, "file.txt", []byte("remote-new"), cfg, out, strings.NewReader("o\n"))
	if err != nil {
		t.Fatalf("deployFile: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "remote-new" {
		t.Errorf("'o' input: file should be overwritten, got %q", data)
	}
	if cfg.DeployedFiles["file.txt"] != hashBytes([]byte("remote-new")) {
		t.Error("'o' input: hash not updated after overwrite")
	}
}

// Fall 4c: Konflikt → Nutzer wählt behalten mit "k" (primäres ASCII-Mnemonic)
func TestDeployFileConflictKeepWithK(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "file.txt")
	writeTestFile(t, dst, "user-modified")

	cfg := &config.Config{
		DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
	}

	out := &strings.Builder{}
	err := deployFile(dst, "file.txt", []byte("remote-new"), cfg, out, strings.NewReader("k\n"))
	if err != nil {
		t.Fatalf("deployFile: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "user-modified" {
		t.Errorf("'k' input: file should be kept, got %q", data)
	}
	if !strings.Contains(out.String(), "KONFLIKT") {
		t.Error("'k' input: conflict should be reported")
	}
}

// Fall 4d: Konflikt → "ü" weiterhin akzeptieren (Backwards-Compat)
func TestDeployFileConflictOverwriteWithUmlaut(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "file.txt")
	writeTestFile(t, dst, "user-modified")

	cfg := &config.Config{
		DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
	}

	out := &strings.Builder{}
	err := deployFile(dst, "file.txt", []byte("remote-new"), cfg, out, strings.NewReader("ü\n"))
	if err != nil {
		t.Fatalf("deployFile: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "remote-new" {
		t.Errorf("'ü' backwards-compat: file should be overwritten, got %q", data)
	}
}

// Fall 4e: Konflikt → "u" weiterhin akzeptieren (Backwards-Compat)
func TestDeployFileConflictOverwriteWithU(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "file.txt")
	writeTestFile(t, dst, "user-modified")

	cfg := &config.Config{
		DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
	}

	out := &strings.Builder{}
	err := deployFile(dst, "file.txt", []byte("remote-new"), cfg, out, strings.NewReader("u\n"))
	if err != nil {
		t.Fatalf("deployFile: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "remote-new" {
		t.Errorf("'u' backwards-compat: file should be overwritten, got %q", data)
	}
}

// Fall 4f: Konflikt → leere Eingabe (Enter) → behalten
func TestDeployFileConflictEmptyInputKeeps(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "file.txt")
	writeTestFile(t, dst, "user-modified")

	cfg := &config.Config{
		DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
	}

	out := &strings.Builder{}
	err := deployFile(dst, "file.txt", []byte("remote-new"), cfg, out, strings.NewReader("\n"))
	if err != nil {
		t.Fatalf("deployFile: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "user-modified" {
		t.Errorf("empty input: file should be kept, got %q", data)
	}
}

// Fall 4g: Konflikt → unbekannte Eingabe → behalten
func TestDeployFileConflictUnknownInputKeeps(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "file.txt")
	writeTestFile(t, dst, "user-modified")

	cfg := &config.Config{
		DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
	}

	out := &strings.Builder{}
	err := deployFile(dst, "file.txt", []byte("remote-new"), cfg, out, strings.NewReader("xyz\n"))
	if err != nil {
		t.Fatalf("deployFile: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "user-modified" {
		t.Errorf("unknown input: file should be kept, got %q", data)
	}
}

// Prompt-Text muss neues Format zeigen
func TestDeployFileConflictPromptText(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "file.txt")
	writeTestFile(t, dst, "user-modified")

	cfg := &config.Config{
		DeployedFiles: map[string]string{"file.txt": hashBytes([]byte("original"))},
	}

	out := &strings.Builder{}
	_ = deployFile(dst, "file.txt", []byte("remote-new"), cfg, out, strings.NewReader("k\n"))

	prompt := out.String()
	if !strings.Contains(prompt, "[o]verwrite") {
		t.Errorf("prompt should contain '[o]verwrite', got: %q", prompt)
	}
	if !strings.Contains(prompt, "[k]eep") {
		t.Errorf("prompt should contain '[k]eep', got: %q", prompt)
	}
}

// Migration: kein stored hash → einfach überschreiben
func TestDeployFileMissingStoredHash(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "file.txt")
	writeTestFile(t, dst, "user-modified")

	cfg := &config.Config{DeployedFiles: map[string]string{}}

	out := &strings.Builder{}
	err := deployFile(dst, "file.txt", []byte("remote-new"), cfg, out, strings.NewReader(""))
	if err != nil {
		t.Fatalf("deployFile: %v", err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "remote-new" {
		t.Errorf("migration: expected overwrite, got %q", data)
	}
	if strings.Contains(out.String(), "KONFLIKT") {
		t.Error("migration should not show conflict")
	}
}
