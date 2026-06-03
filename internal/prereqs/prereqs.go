// Package prereqs detects and (optionally) installs the host binaries required
// to use Claude with forgecrate: the claude CLI, Node.js/npx for the npx-based
// MCP servers, and codegraph for the codegraph MCP server.
package prereqs

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/charmbracelet/huh"
)

// codegraphInstallCmd is the documented codegraph installer. It is a compile-time
// constant so there is no command-injection surface — it is never built from
// user or repository input.
const codegraphInstallCmd = "curl -fsSL https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.sh | sh"

// Tool describes a host binary forgecrate cares about.
type Tool struct {
	Name  string // binary name looked up on PATH
	Hint  string // human-readable install hint shown when missing
	Vital bool   // when true, absence is fatal (e.g. claude)
}

// DefaultTools returns the host tools forgecrate checks for.
func DefaultTools() []Tool {
	return []Tool{
		{Name: "claude", Hint: "https://docs.claude.com/claude-code  (claude CLI ist Pflicht)", Vital: true},
		{Name: "node", Hint: "https://nodejs.org  (für npx-basierte MCP-Server)"},
		{Name: "npx", Hint: "https://nodejs.org  (npx fehlt — alle npx-MCP-Server schlagen sonst fehl)"},
		{Name: "codegraph", Hint: codegraphInstallCmd},
	}
}

// Checker detects and installs prerequisites.
type Checker struct {
	Out    io.Writer // progress output; nil = silent
	DryRun bool      // print actions instead of executing them
	Assume bool      // treat confirmations as "yes" (non-interactive / --yes)
}

func (c Checker) out() io.Writer {
	if c.Out == nil {
		return io.Discard
	}
	return c.Out
}

func has(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Detect partitions tools into those present and those missing on PATH.
func (c Checker) Detect(tools []Tool) (present, missing []Tool) {
	for _, t := range tools {
		if has(t.Name) {
			present = append(present, t)
		} else {
			missing = append(missing, t)
		}
	}
	return present, missing
}

// EnsureClaude returns an error if the claude CLI is not on PATH.
func (c Checker) EnsureClaude() error {
	if has("claude") {
		return nil
	}
	return fmt.Errorf("claude CLI nicht gefunden — Installation: https://docs.claude.com/claude-code")
}

// EnsureCodegraph installs codegraph via its documented install.sh, but only
// after explicit confirmation (unless Assume or DryRun). It is a no-op when
// codegraph is already present.
func (c Checker) EnsureCodegraph() error {
	out := c.out()
	if has("codegraph") {
		_, _ = fmt.Fprintln(out, "🔵 codegraph bereits installiert")
		return nil
	}

	if c.DryRun {
		_, _ = fmt.Fprintf(out, "🟡 dry-run codegraph:  %s\n", codegraphInstallCmd)
		return nil
	}

	if !c.Assume {
		var confirm bool
		err := huh.NewConfirm().
			Title("codegraph nicht gefunden. Offizielles install.sh ausführen?").
			Description(codegraphInstallCmd).
			Affirmative("Ja, installieren").
			Negative("Nein, überspringen").
			Value(&confirm).
			Run()
		if err != nil {
			return err
		}
		if !confirm {
			_, _ = fmt.Fprintln(out, "⏭️  codegraph übersprungen")
			return nil
		}
	}

	_, _ = fmt.Fprintf(out, "⚠️  Führe Remote-Installer aus:  %s\n", codegraphInstallCmd)
	cmd := exec.Command("sh", "-c", codegraphInstallCmd)
	if cmdOut, err := cmd.CombinedOutput(); err != nil {
		_, _ = fmt.Fprintf(out, "%s\n", cmdOut)
		return fmt.Errorf("codegraph-Installation fehlgeschlagen: %w", err)
	}

	if !has("codegraph") {
		return fmt.Errorf("codegraph nach Installation nicht auf PATH — ggf. Shell neu laden")
	}
	_, _ = fmt.Fprintln(out, "✅ codegraph installiert")
	return nil
}
