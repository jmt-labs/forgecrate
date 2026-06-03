package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	hook.AddCommand(newHookPreToolCmd())
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
		fmt.Fprintf(&sb, "Recherche-Pflicht: einmal pro Session vor dem ersten Edit/Write WebSearch/context7/fetch nutzen — nicht raten (Warnung via pre-tool Hook).\n")
	}
	return sb.String(), nil
}

func newHookRequireResearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "require-research",
		Short: "Warnt bei Edit/Write/MultiEdit (und schreibender Bash bei force-research), wenn in der Session noch nicht recherchiert wurde",
		RunE: func(cmd *cobra.Command, args []string) error {
			if out := requireResearchOutput(os.Stdin, "."); out != "" {
				fmt.Print(out)
			}
			return nil
		},
	}
}

const researchBlockMessage = "Recherche-Empfehlung: Vor Edit/Write/MultiEdit mindestens ein Recherche-Tool (WebSearch, WebFetch, mcp__fetch__*, mcp__context7__*) nutzen — nicht raten. Einmal pro Session genügt, danach sind weitere Edits frei. Verzicht via Flavor no-research."

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

// requireResearchOutput liest das PreToolUse-stdin-JSON und gibt bei fehlender
// Recherche eine Warnung aus, sonst einen leeren String. Fail-open bei Lesefehlern.
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

	warn, reason := researchDecision(cfg, transcript, in.ToolName, in.ToolInput.Command)
	if !warn {
		return ""
	}
	out, err := json.Marshal(map[string]any{
		"hookSpecificOutput": map[string]string{
			"hookEventName":     "PreToolUse",
			"additionalContext": reason,
		},
	})
	if err != nil {
		return ""
	}
	return string(out)
}

// researchDecision entscheidet, ob bei einem Tool-Aufruf eine Recherche-Warnung ausgegeben
// wird, weil in der Session noch kein Recherche-Tool genutzt wurde. no-research deaktiviert.
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

	if transcriptHasResearchAnywhere(transcript) {
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

// transcriptHasResearchAnywhere prüft, ob irgendwo in der Session (gesamtes
// Transcript) ein assistant-tool_use mit einem Recherche-Tool vorkommt. Einmal pro
// Session genügt — Folge-Edits werden nicht erneut geblockt. Robust gegen kaputte Zeilen.
func transcriptHasResearchAnywhere(transcript []byte) bool {
	lines := bytes.Split(transcript, []byte("\n"))

	for _, raw := range lines {
		raw = bytes.TrimSpace(raw)
		if len(raw) == 0 {
			continue
		}
		var tl transcriptLine
		if err := json.Unmarshal(raw, &tl); err != nil {
			continue
		}
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

func newHookPreToolCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pre-tool",
		Short: "Prüft destruktive Tool-Aufrufe vor der Ausführung",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, _ := io.ReadAll(os.Stdin)
			var in preToolInput
			_ = json.Unmarshal(data, &in)
			if in.ToolName == "" {
				in.ToolName = os.Getenv("CLAUDE_TOOL_NAME")
			}
			if in.ToolInput.Command == "" {
				in.ToolInput.Command = os.Getenv("TOOL_INPUT")
			}
			branch, _ := currentBranch()
			if out := preToolOutput(branch, in.ToolName, in.ToolInput.Command); out != "" {
				fmt.Print(out)
			}
			return nil
		},
	}
}

func currentBranch() (string, error) {
	out, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func isMainBranch(branch string) bool {
	return branch == "main" || branch == "master"
}

var reDestructiveBash = []*regexp.Regexp{
	regexp.MustCompile(`(^|[;&|]|\brun\b)\s*git\s+commit\b`),
	regexp.MustCompile(`git\s+push\s+.*(-f\b|--force\b)`),
	regexp.MustCompile(`git\s+push\b.*\b(main|master)\b`),
	regexp.MustCompile(`git\s+reset\s+--hard\b`),
	regexp.MustCompile(`git\s+clean\s+.*-[a-zA-Z]*f`),
	regexp.MustCompile(`>>?\s*[^/\s][^\s]*`),
}

func isDestructiveBash(cmd string) string {
	patterns := []struct {
		re  *regexp.Regexp
		msg string
	}{
		{reDestructiveBash[0], "git commit"},
		{reDestructiveBash[1], "git push --force"},
		{reDestructiveBash[2], "git push ... main/master"},
		{reDestructiveBash[3], "git reset --hard"},
		{reDestructiveBash[4], "git clean -f"},
		{reDestructiveBash[5], "Schreib-Redirektion"},
	}
	for _, p := range patterns {
		if p.re.MatchString(cmd) {
			if p.msg == "Schreib-Redirektion" && strings.Contains(cmd, "/tmp/") {
				continue
			}
			return p.msg
		}
	}
	return ""
}

func preToolOutput(branch, toolName, toolInput string) string {
	onMain := isMainBranch(branch)

	switch toolName {
	case "Edit", "Write", "MultiEdit":
		if onMain {
			out, _ := json.Marshal(map[string]any{
				"hookSpecificOutput": map[string]string{
					"hookEventName":     "PreToolUse",
					"additionalContext": "Warnung: Direkte Änderungen auf main. Branch anlegen empfohlen: git checkout -b feat/<thema>",
				},
			})
			return string(out)
		}
	case "Bash":
		destructive := isDestructiveBash(toolInput)
		if destructive == "" {
			return ""
		}
		msg := "Warnung: destruktiver Befehl erkannt (" + destructive + ")."
		if onMain {
			msg += " Direkt auf main — Branch anlegen empfohlen."
		} else {
			msg += " Mit Bedacht verwenden."
		}
		out, _ := json.Marshal(map[string]any{
			"hookSpecificOutput": map[string]string{
				"hookEventName":     "PreToolUse",
				"additionalContext": msg,
			},
		})
		return string(out)
	}
	return ""
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
