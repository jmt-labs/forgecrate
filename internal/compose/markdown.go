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
	cleaned := make([]string, len(layers))
	for i, l := range layers {
		cleaned[i] = stripWrapperMarkers(l)
	}
	generated := strings.Join(cleaned, "\n\n")
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

func stripWrapperMarkers(s string) string {
	for _, marker := range []string{generatedBegin, generatedEnd, customBegin, customEnd} {
		s = strings.ReplaceAll(s, marker+"\n", "")
		s = strings.ReplaceAll(s, marker, "")
	}
	return strings.TrimSpace(s)
}

func extractCustom(existing string) string {
	start := strings.Index(existing, customBegin)
	end := strings.Index(existing, customEnd)

	if start == -1 || end == -1 || end <= start {
		// Keine Marker oder ungültige Marker: existierende Datei als Custom behandeln
		if strings.TrimSpace(existing) != "" {
			return existing + "\n"
		}
		return ""
	}

	content := existing[start+len(customBegin) : end]
	return strings.TrimLeft(content, "\n")
}
