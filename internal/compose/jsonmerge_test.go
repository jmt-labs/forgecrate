package compose_test

import (
	"encoding/json"
	"testing"

	"github.com/jmt-labs/claude-setup/internal/compose"
)

func TestDeepMergeJSON(t *testing.T) {
	base := `{"hooks":{"UserPromptSubmit":[{"matcher":"","hooks":[{"type":"command","command":"bash a.sh"}]}]},"permissions":{"allow":["Bash"]}}`
	override := `{"permissions":{"allow":["Bash","Edit"]},"model":"claude-opus-4-7"}`

	result, err := compose.DeepMergeJSON(base, override)
	if err != nil {
		t.Fatalf("DeepMergeJSON: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(result), &m); err != nil {
		t.Fatalf("result not valid JSON: %v", err)
	}

	if m["model"] != "claude-opus-4-7" {
		t.Errorf("model: got %v", m["model"])
	}

	perms := m["permissions"].(map[string]any)
	allow := perms["allow"].([]any)
	if len(allow) != 2 {
		t.Errorf("allow: got %d elements, want 2", len(allow))
	}

	hooks := m["hooks"].(map[string]any)
	if hooks["UserPromptSubmit"] == nil {
		t.Error("hooks.UserPromptSubmit missing from merge result")
	}
}

func TestDeepMergeJSONEmpty(t *testing.T) {
	result, err := compose.DeepMergeJSON(`{"a":1}`, `{}`)
	if err != nil {
		t.Fatalf("DeepMergeJSON: %v", err)
	}
	var m map[string]any
	json.Unmarshal([]byte(result), &m)
	if m["a"] != float64(1) {
		t.Errorf("a: got %v", m["a"])
	}
}

func TestDeepMergeJSONInvalidBase(t *testing.T) {
	_, err := compose.DeepMergeJSON(`{invalid}`, `{}`)
	if err == nil {
		t.Error("expected error for invalid base JSON")
	}
}
