package compose

import (
	"strings"
)

const (
	generatedBegin = "<!-- GENERATED:BEGIN -->"
	generatedEnd   = "<!-- GENERATED:END -->"
	customBegin    = "<!-- CUSTOM:BEGIN -->"
	customEnd      = "<!-- CUSTOM:END -->"
)

// MergeMarkdown compositioniert mehrere Markdown-Layer zu einem String.
// existing ist der aktuelle Dateiinhalt (leer bei init).
func MergeMarkdown(layers []string, existing string) string {
	generated := strings.Join(layers, "\n\n")
	custom := extractCustom(existing)

	var b strings.Builder
	b.WriteString(generatedBegin + "\n")
	b.WriteString(generated + "\n")
	b.WriteString(generatedEnd + "\n")
	b.WriteString("\n")
	b.WriteString(customBegin + "\n")
	b.WriteString(custom)
	b.WriteString(customEnd + "\n")
	return b.String()
}

func extractCustom(existing string) string {
	start := strings.Index(existing, customBegin)
	end := strings.Index(existing, customEnd)

	if start == -1 || end == -1 {
		// Keine Marker: existierende Datei als Custom behandeln
		if strings.TrimSpace(existing) != "" {
			return existing + "\n"
		}
		return ""
	}

	content := existing[start+len(customBegin) : end]
	return strings.TrimLeft(content, "\n")
}
