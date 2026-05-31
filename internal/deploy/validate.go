package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmt-labs/forgecrate/internal/config"
)

// validateSelection prüft, ob das gewählte Profil und alle Flavors im Source-Repo
// existieren. Validiert wird gegen den tatsächlichen Katalog (profiles/, flavors/);
// ist ein Katalog leer oder nicht vorhanden, entfällt die Prüfung für diese Kategorie.
func validateSelection(sourceDir string, cfg config.Config) error {
	profiles, err := dirNames(filepath.Join(sourceDir, "profiles"))
	if err != nil {
		return err
	}
	if len(profiles) > 0 {
		if cfg.Profile == "" {
			return fmt.Errorf("kein Profil angegeben — verfügbar: %s", strings.Join(profiles, ", "))
		}
		if !contains(profiles, cfg.Profile) {
			return fmt.Errorf("unbekanntes Profil %q%s", cfg.Profile, availabilityMsg(cfg.Profile, profiles))
		}
	}

	flavors, err := dirNames(filepath.Join(sourceDir, "flavors"))
	if err != nil {
		return err
	}
	if len(flavors) > 0 {
		for _, f := range cfg.Flavors {
			if !contains(flavors, f) {
				return fmt.Errorf("unbekannter Flavor %q%s", f, availabilityMsg(f, flavors))
			}
		}
	}
	return nil
}

// availabilityMsg ergänzt einen optionalen „meintest du …?"-Hinweis und listet
// die verfügbaren Optionen auf.
func availabilityMsg(input string, candidates []string) string {
	if s := suggest(input, candidates); s != "" {
		return fmt.Sprintf(" — meintest du %q? Verfügbar: %s", s, strings.Join(candidates, ", "))
	}
	return fmt.Sprintf(" — verfügbar: %s", strings.Join(candidates, ", "))
}

// dirNames liefert die Namen aller Unterverzeichnisse von dir, alphabetisch sortiert.
// Existiert dir nicht, wird eine leere Liste ohne Fehler zurückgegeben.
func dirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// suggest liefert den ähnlichsten Kandidaten, sofern dessen Levenshtein-Distanz
// innerhalb des Schwellwerts max(2, len(input)/2) liegt — sonst einen leeren String.
func suggest(input string, candidates []string) string {
	threshold := len(input) / 2
	if threshold < 2 {
		threshold = 2
	}
	best := ""
	bestDist := -1
	for _, c := range candidates {
		d := levenshtein(input, c)
		if d <= threshold && (bestDist == -1 || d < bestDist) {
			bestDist = d
			best = c
		}
	}
	return best
}

// levenshtein berechnet die Edit-Distanz zwischen a und b (zeilenweise DP).
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			curr[j] = min(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}
