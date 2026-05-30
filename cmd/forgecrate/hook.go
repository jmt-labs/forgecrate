package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jmt-labs/forgecrate/internal/config"
	"github.com/spf13/cobra"
)

func newHookCmd() *cobra.Command {
	hook := &cobra.Command{
		Use:   "hook",
		Short: "Hook-Hilfsprogramme für Claude Code",
	}
	hook.AddCommand(newHookPromptSubmitCmd())
	hook.AddCommand(newHookRequireResearchCmd())
	return hook
}

func newHookPromptSubmitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prompt-submit",
		Short: "Gibt die aktive forgecrate-Konfiguration aus (für prompt-submit Hook)",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := promptSubmitOutput(".")
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		},
	}
}

// readForgecrateConfig liest .forgecrate.yaml mit Fallback auf .claude-setup.yaml.
func readForgecrateConfig(dir string) (config.Config, error) {
	cfg, err := config.Read(filepath.Join(dir, ".forgecrate.yaml"))
	if err != nil {
		cfg, err = config.Read(filepath.Join(dir, ".claude-setup.yaml"))
	}
	return cfg, err
}

func promptSubmitOutput(dir string) (string, error) {
	cfg, err := readForgecrateConfig(dir)

	var profile, flavors string
	if err != nil {
		profile = "unbekannt"
		flavors = "keine"
	} else {
		profile = cfg.Profile
		if profile == "" {
			profile = "unbekannt"
		}
		if len(cfg.Flavors) > 0 {
			flavors = strings.Join(cfg.Flavors, ", ")
		} else {
			flavors = "keine"
		}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "## forgecrate — Aktive Konfiguration\n")
	fmt.Fprintf(&sb, "Profil: %s | Flavors: %s\n", profile, flavors)
	fmt.Fprintln(&sb)
	fmt.Fprintf(&sb, "Pflicht-Skills: brainstorming → tdd → verification-before-completion | debugging bei Bugs\n")
	if !cfg.HasFlavor("no-research") {
		fmt.Fprintf(&sb, "Recherche-Pflicht (erzwungen): vor jedem Edit/Write WebSearch/context7/fetch nutzen — nicht raten (Block via pre-tool Hook).\n")
	}
	return sb.String(), nil
}

func newHookRequireResearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "require-research",
		Short: "Blockiert Edit/Write/MultiEdit (und schreibende Bash bei force-research), bis im Turn recherchiert wurde",
		RunE: func(cmd *cobra.Command, args []string) error {
			if out := requireResearchOutput(os.Stdin, "."); out != "" {
				fmt.Print(out)
			}
			return nil
		},
	}
}

const researchBlockMessage = "Recherche-Pflicht: Vor Edit/Write/MultiEdit muss in diesem Turn mindestens ein Recherche-Tool (WebSearch, WebFetch, mcp__fetch__*, mcp__context7__*) genutzt worden sein. Recherchiere zuerst die relevante Doku/Best Practice, dann editiere. Bewusster Verzicht: Flavor no-research aktivieren."

// preToolInput ist das stdin-JSON, das ein PreToolUse-Hook von Claude Code erhält.
type preToolInput struct {
	ToolName       string `json:"tool_name"`
	TranscriptPath string `json:"transcript_path"`
	ToolInput      struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

// transcriptLine ist eine Zeile der Transcript-JSONL (nur benötigte Felder).
type transcriptLine struct {
	Type    string `json:"type"`
	Message struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type contentBlock struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type hookSpecificOutput struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision"`
	PermissionDecisionReason string `json:"permissionDecisionReason"`
}

type hookOutput struct {
	HookSpecificOutput hookSpecificOutput `json:"hookSpecificOutput"`
}

// requireResearchOutput liest das PreToolUse-stdin-JSON und gibt bei einem Block das
// deny-JSON zurück, sonst einen leeren String. Fail-open bei jedem Lesefehler.
func requireResearchOutput(r io.Reader, dir string) string {
	data, err := io.ReadAll(r)
	if err != nil {
		return ""
	}
	var in preToolInput
	if err := json.Unmarshal(data, &in); err != nil {
		return ""
	}

	cfg, _ := readForgecrateConfig(dir)

	// Ohne lesbaren Transcript kann keine Recherche nachgewiesen werden — fail-open,
	// um Dauersperren bei fehlendem/kaputtem transcript_path zu vermeiden.
	if in.TranscriptPath == "" {
		return ""
	}
	transcript, err := os.ReadFile(in.TranscriptPath)
	if err != nil {
		return ""
	}

	block, reason := researchDecision(cfg, transcript, in.ToolName, in.ToolInput.Command)
	if !block {
		return ""
	}
	out, err := json.Marshal(hookOutput{HookSpecificOutput: hookSpecificOutput{
		HookEventName:            "PreToolUse",
		PermissionDecision:       "deny",
		PermissionDecisionReason: reason,
	}})
	if err != nil {
		return ""
	}
	return string(out)
}

// researchDecision entscheidet, ob ein Tool-Aufruf blockiert wird, weil im aktuellen
// User-Turn noch kein Recherche-Tool genutzt wurde. no-research deaktiviert den Block.
func researchDecision(cfg config.Config, transcript []byte, toolName, bashCmd string) (bool, string) {
	if cfg.HasFlavor("no-research") {
		return false, ""
	}

	switch toolName {
	case "Edit", "Write", "MultiEdit":
		// immer gegated (siehe unten)
	case "Bash":
		if !cfg.HasFlavor("force-research") || !bashWrites(bashCmd) {
			return false, ""
		}
	default:
		return false, ""
	}

	if transcriptHasResearchSinceLastUser(transcript) {
		return false, ""
	}
	return true, researchBlockMessage
}

func isResearchTool(name string) bool {
	switch name {
	case "WebSearch", "WebFetch":
		return true
	}
	return strings.HasPrefix(name, "mcp__fetch__") || strings.HasPrefix(name, "mcp__context7__")
}

// transcriptHasResearchSinceLastUser prüft, ob nach dem letzten user-Eintrag ein
// assistant-tool_use mit einem Recherche-Tool vorkommt. Robust gegen kaputte Zeilen.
func transcriptHasResearchSinceLastUser(transcript []byte) bool {
	lines := bytes.Split(transcript, []byte("\n"))
	parsed := make([]transcriptLine, len(lines))
	valid := make([]bool, len(lines))
	lastUser := -1

	for i, raw := range lines {
		raw = bytes.TrimSpace(raw)
		if len(raw) == 0 {
			continue
		}
		var tl transcriptLine
		if err := json.Unmarshal(raw, &tl); err != nil {
			continue
		}
		parsed[i] = tl
		valid[i] = true
		if tl.Type == "user" || tl.Message.Role == "user" {
			lastUser = i
		}
	}

	for i := lastUser + 1; i < len(lines); i++ {
		if !valid[i] {
			continue
		}
		tl := parsed[i]
		if tl.Type != "assistant" && tl.Message.Role != "assistant" {
			continue
		}
		var blocks []contentBlock
		if err := json.Unmarshal(tl.Message.Content, &blocks); err != nil {
			continue
		}
		for _, b := range blocks {
			if b.Type == "tool_use" && isResearchTool(b.Name) {
				return true
			}
		}
	}
	return false
}

var (
	reSedInplace = regexp.MustCompile(`\bsed\b[^|;&]*\s-i`)
	reTee        = regexp.MustCompile(`\btee\b`)
	reDdOf       = regexp.MustCompile(`\bdd\b[^|;&]*\bof=`)
	reRedirect   = regexp.MustCompile(`>>?\s*([^\s>][^\s]*)`)
)

// bashWrites erkennt heuristisch, ob ein Bash-Befehl in eine versionierte Datei
// schreibt (Redirect außerhalb /tmp, sed -i, tee, dd of=).
func bashWrites(cmd string) bool {
	if reSedInplace.MatchString(cmd) || reTee.MatchString(cmd) || reDdOf.MatchString(cmd) {
		return true
	}
	for _, m := range reRedirect.FindAllStringSubmatch(cmd, -1) {
		if !strings.HasPrefix(m[1], "/tmp/") {
			return true
		}
	}
	return false
}
