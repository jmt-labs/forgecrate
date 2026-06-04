package extensions

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// CheckNodeVersion prüft ob node >= minMajor vorhanden ist.
// Gibt Fehler zurück wenn node fehlt oder die Major-Version < minMajor ist.
func CheckNodeVersion(minMajor int) error {
	out, err := exec.Command("node", "--version").Output()
	if err != nil {
		return fmt.Errorf("node nicht gefunden — Node.js %d+ erforderlich (https://nodejs.org)", minMajor)
	}

	version := strings.TrimSpace(string(out))
	version = strings.TrimPrefix(version, "v")
	parts := strings.SplitN(version, ".", 2)
	if len(parts) == 0 {
		return fmt.Errorf("node --version unbekanntes Format: %q", string(out))
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("node --version unbekanntes Format: %q", string(out))
	}

	if major < minMajor {
		return fmt.Errorf("node v%s gefunden, aber Node.js %d+ erforderlich (https://nodejs.org)", strings.TrimSpace(string(out)), minMajor)
	}

	return nil
}
