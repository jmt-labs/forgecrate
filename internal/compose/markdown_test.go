package compose_test

import (
	"strings"
	"testing"

	"github.com/markus/claude-setup/internal/compose"
)

func TestMergeMarkdownInit(t *testing.T) {
	layers := []string{"# Base\n\nBase content.", "## Profile\n\nProfile content."}
	result := compose.MergeMarkdown(layers, "")

	want := "<!-- GENERATED:BEGIN -->\n# Base\n\nBase content.\n\n## Profile\n\nProfile content.\n<!-- GENERATED:END -->\n\n<!-- CUSTOM:BEGIN -->\n<!-- CUSTOM:END -->\n"
	if result != want {
		t.Errorf("got:\n%q\nwant:\n%q", result, want)
	}
}

func TestMergeMarkdownPreservesCustom(t *testing.T) {
	existing := "<!-- GENERATED:BEGIN -->\n# Old\n<!-- GENERATED:END -->\n\n<!-- CUSTOM:BEGIN -->\n# My custom section\n<!-- CUSTOM:END -->\n"
	layers := []string{"# New Base"}
	result := compose.MergeMarkdown(layers, existing)

	if !strings.Contains(result, "# My custom section") {
		t.Error("custom section was lost")
	}
	if !strings.Contains(result, "# New Base") {
		t.Error("new generated content missing")
	}
}

func TestMergeMarkdownNoExistingMarkers(t *testing.T) {
	existing := "# Handwritten file\n\nNo markers here."
	layers := []string{"# Base"}
	result := compose.MergeMarkdown(layers, existing)

	if !strings.Contains(result, "# Handwritten file") {
		t.Error("existing content without markers was lost")
	}
}
