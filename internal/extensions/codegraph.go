package extensions

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// InitCodegraph initialisiert den codegraph-Index im destDir wenn das Binary vorhanden ist.
// Warnt wenn das Binary fehlt. Überspringt wenn .codegraph/ bereits existiert (idempotent).
func InitCodegraph(destDir string, codegraphBin string, out io.Writer) error {
	if codegraphBin == "" {
		codegraphBin = "codegraph"
	}
	if out == nil {
		out = io.Discard
	}

	binPath, err := exec.LookPath(codegraphBin)
	if err != nil {
		_, _ = fmt.Fprintf(out, "🟡 codegraph nicht gefunden — Binary installieren damit der Index aufgebaut wird\n")
		return nil
	}

	dotCodegraph := filepath.Join(destDir, ".codegraph")
	if _, err := os.Stat(dotCodegraph); err == nil {
		_, _ = fmt.Fprintf(out, "🔵 codegraph: .codegraph/ existiert bereits, Init wird übersprungen\n")
		return nil
	}

	cmd := exec.Command(binPath, "init", "-i")
	cmd.Dir = destDir
	if cmdOut, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("codegraph init -i: %w\n%s", err, cmdOut)
	}

	return nil
}
