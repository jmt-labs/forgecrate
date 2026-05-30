package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmt-labs/forgecrate/internal/config"
)

func writeYAML(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, ".forgecrate.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func TestPromptSubmitOutput_BlockList(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, `profile: backend
flavors:
  - tdd
  - strict-review
`)
	out, err := promptSubmitOutput(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Profil: backend") {
		t.Errorf("expected 'Profil: backend', got: %s", out)
	}
	if !strings.Contains(out, "tdd") {
		t.Errorf("expected 'tdd' in output, got: %s", out)
	}
	if !strings.Contains(out, "strict-review") {
		t.Errorf("expected 'strict-review' in output, got: %s", out)
	}
}

func TestPromptSubmitOutput_InlineList(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, `profile: backend
flavors: [tdd, strict-review]
`)
	out, err := promptSubmitOutput(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Profil: backend") {
		t.Errorf("expected 'Profil: backend', got: %s", out)
	}
	if !strings.Contains(out, "tdd") {
		t.Errorf("expected 'tdd' in output, got: %s", out)
	}
	if !strings.Contains(out, "strict-review") {
		t.Errorf("expected 'strict-review' in output, got: %s", out)
	}
}

func TestPromptSubmitOutput_CommentIgnored(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, `# flavors: [this-should-not-appear]
profile: backend
flavors:
  - tdd
`)
	out, err := promptSubmitOutput(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "this-should-not-appear") {
		t.Errorf("comment value must not appear in output, got: %s", out)
	}
}

func TestPromptSubmitOutput_MissingFile(t *testing.T) {
	dir := t.TempDir()
	out, err := promptSubmitOutput(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "unbekannt") {
		t.Errorf("expected 'unbekannt' for missing config, got: %s", out)
	}
	if !strings.Contains(out, "keine") {
		t.Errorf("expected 'keine' for missing config, got: %s", out)
	}
}

func TestPromptSubmitOutput_FallbackFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude-setup.yaml")
	if err := os.WriteFile(path, []byte("profile: frontend\nflavors:\n  - github\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	out, err := promptSubmitOutput(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Profil: frontend") {
		t.Errorf("expected 'Profil: frontend', got: %s", out)
	}
}

func TestPromptSubmitOutput_ContainsSkillsLine(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "profile: backend\nflavors:\n  - tdd\n")
	out, err := promptSubmitOutput(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Pflicht-Skills") {
		t.Errorf("expected 'Pflicht-Skills' line, got: %s", out)
	}
	if !strings.Contains(out, "brainstorming") {
		t.Errorf("expected 'brainstorming' in output, got: %s", out)
	}
}

func TestPromptSubmitOutput_ResearchReminderDefault(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "profile: backend\nflavors:\n  - tdd\n")
	out, err := promptSubmitOutput(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Recherche-Pflicht") {
		t.Errorf("expected research reminder by default, got: %s", out)
	}
	if !strings.Contains(out, "WebSearch") {
		t.Errorf("expected 'WebSearch' in reminder, got: %s", out)
	}
}

func TestPromptSubmitOutput_ResearchReminderOptOut(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "profile: backend\nflavors:\n  - tdd\n  - no-research\n")
	out, err := promptSubmitOutput(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "Recherche-Pflicht") {
		t.Errorf("research reminder must NOT appear when no-research flavor is active, got: %s", out)
	}
}

func TestPromptSubmitOutput_ResearchReminderMissingConfig(t *testing.T) {
	dir := t.TempDir()
	out, err := promptSubmitOutput(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Recherche-Pflicht") {
		t.Errorf("expected research reminder when config is missing (default on), got: %s", out)
	}
}

// --- require-research / researchDecision ---

func userLine() string {
	return `{"type":"user","message":{"role":"user","content":"tu etwas"}}`
}

func toolUseLine(name string) string {
	return `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"` + name + `","input":{}}]}}`
}

func textLine(text string) string {
	return `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"` + text + `"}]}}`
}

func transcript(lines ...string) []byte {
	return []byte(strings.Join(lines, "\n") + "\n")
}

func cfgWith(flavors ...string) config.Config {
	return config.Config{Profile: "backend", Flavors: flavors}
}

func TestIsResearchTool(t *testing.T) {
	research := []string{"WebSearch", "WebFetch", "mcp__fetch__fetch", "mcp__context7__query-docs", "mcp__context7__resolve-library-id"}
	for _, name := range research {
		if !isResearchTool(name) {
			t.Errorf("expected %q to be a research tool", name)
		}
	}
	notResearch := []string{"Grep", "Glob", "Read", "Edit", "Bash", "mcp__github__search_code", "mcp__memory__read_graph", ""}
	for _, name := range notResearch {
		if isResearchTool(name) {
			t.Errorf("expected %q NOT to be a research tool", name)
		}
	}
}

func TestResearchDecision(t *testing.T) {
	tests := []struct {
		name       string
		cfg        config.Config
		transcript []byte
		toolName   string
		bashCmd    string
		wantBlock  bool
	}{
		{
			name:       "no-research flavor disables block",
			cfg:        cfgWith("no-research"),
			transcript: transcript(userLine(), textLine("kein Research")),
			toolName:   "Edit",
			wantBlock:  false,
		},
		{
			name:       "Read is never blocked",
			cfg:        cfgWith(),
			transcript: transcript(userLine(), textLine("kein Research")),
			toolName:   "Read",
			wantBlock:  false,
		},
		{
			name:       "Bash without force-research is never blocked",
			cfg:        cfgWith(),
			transcript: transcript(userLine(), textLine("kein Research")),
			toolName:   "Bash",
			bashCmd:    "sed -i 's/a/b/' file.go",
			wantBlock:  false,
		},
		{
			name:       "Edit without prior research is blocked",
			cfg:        cfgWith(),
			transcript: transcript(userLine(), textLine("kein Research")),
			toolName:   "Edit",
			wantBlock:  true,
		},
		{
			name:       "Edit after WebSearch in turn is allowed",
			cfg:        cfgWith(),
			transcript: transcript(userLine(), toolUseLine("WebSearch")),
			toolName:   "Write",
			wantBlock:  false,
		},
		{
			name:       "research in an earlier turn of the session still counts",
			cfg:        cfgWith(),
			transcript: transcript(userLine(), toolUseLine("WebSearch"), userLine(), textLine("nichts")),
			toolName:   "MultiEdit",
			wantBlock:  false,
		},
		{
			name:       "single research early in session unlocks later edits without re-research",
			cfg:        cfgWith(),
			transcript: transcript(userLine(), toolUseLine("mcp__context7__query-docs"), userLine(), textLine("egal"), userLine(), textLine("immer noch egal")),
			toolName:   "Edit",
			wantBlock:  false,
		},
		{
			name:       "context7 tool_use satisfies requirement",
			cfg:        cfgWith(),
			transcript: transcript(userLine(), toolUseLine("mcp__context7__query-docs")),
			toolName:   "Edit",
			wantBlock:  false,
		},
		{
			name:       "fetch mcp tool_use satisfies requirement",
			cfg:        cfgWith(),
			transcript: transcript(userLine(), toolUseLine("mcp__fetch__fetch")),
			toolName:   "Edit",
			wantBlock:  false,
		},
		{
			name:       "force-research blocks writing bash without research",
			cfg:        cfgWith("force-research"),
			transcript: transcript(userLine(), textLine("nichts")),
			toolName:   "Bash",
			bashCmd:    "sed -i 's/a/b/' file.go",
			wantBlock:  true,
		},
		{
			name:       "force-research allows writing bash after research",
			cfg:        cfgWith("force-research"),
			transcript: transcript(userLine(), toolUseLine("WebSearch")),
			toolName:   "Bash",
			bashCmd:    "echo hi > file.go",
			wantBlock:  false,
		},
		{
			name:       "force-research ignores non-writing bash",
			cfg:        cfgWith("force-research"),
			transcript: transcript(userLine(), textLine("nichts")),
			toolName:   "Bash",
			bashCmd:    "ls -la",
			wantBlock:  false,
		},
		{
			name:       "broken jsonl lines mixed with valid research",
			cfg:        cfgWith(),
			transcript: []byte("not json\n" + userLine() + "\n{garbage\n" + toolUseLine("WebSearch") + "\n"),
			toolName:   "Edit",
			wantBlock:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, reason := researchDecision(tt.cfg, tt.transcript, tt.toolName, tt.bashCmd)
			if block != tt.wantBlock {
				t.Errorf("researchDecision block = %v, want %v (reason: %q)", block, tt.wantBlock, reason)
			}
			if block && !strings.Contains(reason, "Recherche") {
				t.Errorf("expected reason to mention 'Recherche', got: %q", reason)
			}
		})
	}
}

func TestRequireResearchOutput(t *testing.T) {
	dir := t.TempDir()
	noResearch := filepath.Join(dir, "no_research.jsonl")
	if err := os.WriteFile(noResearch, transcript(userLine(), textLine("ok")), 0644); err != nil {
		t.Fatal(err)
	}
	withResearch := filepath.Join(dir, "with_research.jsonl")
	if err := os.WriteFile(withResearch, transcript(userLine(), toolUseLine("WebSearch")), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("Edit without research blocks", func(t *testing.T) {
		in := `{"tool_name":"Edit","transcript_path":"` + noResearch + `"}`
		out := requireResearchOutput(strings.NewReader(in), dir)
		if !strings.Contains(out, `"permissionDecision":"deny"`) {
			t.Errorf("expected deny output, got: %q", out)
		}
	})
	t.Run("Edit after research allowed", func(t *testing.T) {
		in := `{"tool_name":"Edit","transcript_path":"` + withResearch + `"}`
		if out := requireResearchOutput(strings.NewReader(in), dir); out != "" {
			t.Errorf("expected empty output, got: %q", out)
		}
	})
	t.Run("missing transcript file fails open", func(t *testing.T) {
		in := `{"tool_name":"Edit","transcript_path":"` + filepath.Join(dir, "nope.jsonl") + `"}`
		if out := requireResearchOutput(strings.NewReader(in), dir); out != "" {
			t.Errorf("expected fail-open empty output for missing transcript, got: %q", out)
		}
	})
	t.Run("empty transcript path fails open", func(t *testing.T) {
		in := `{"tool_name":"Edit","transcript_path":""}`
		if out := requireResearchOutput(strings.NewReader(in), dir); out != "" {
			t.Errorf("expected fail-open empty output for empty path, got: %q", out)
		}
	})
	t.Run("invalid stdin json fails open", func(t *testing.T) {
		if out := requireResearchOutput(strings.NewReader("{not json"), dir); out != "" {
			t.Errorf("expected fail-open empty output for invalid json, got: %q", out)
		}
	})
}

func TestBashWrites(t *testing.T) {
	writing := []string{
		"sed -i 's/a/b/' f.go",
		"echo hi > f.go",
		"echo hi >> f.go",
		"cat a | tee f.go",
		"dd if=/dev/zero of=f.bin",
	}
	for _, cmd := range writing {
		if !bashWrites(cmd) {
			t.Errorf("expected %q to be detected as writing", cmd)
		}
	}
	nonWriting := []string{
		"ls -la",
		"git status",
		"echo hi > /tmp/scratch.txt",
		"grep foo bar.go",
	}
	for _, cmd := range nonWriting {
		if bashWrites(cmd) {
			t.Errorf("expected %q NOT to be detected as writing", cmd)
		}
	}
}
