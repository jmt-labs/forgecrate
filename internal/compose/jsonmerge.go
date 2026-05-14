package compose

import (
	"encoding/json"
	"fmt"
)

// DeepMergeJSON merged zwei JSON-Objekte. Override-Werte gewinnen.
// Arrays werden ersetzt (nicht gemergt). Objekte werden rekursiv gemergt.
func DeepMergeJSON(base, override string) (string, error) {
	var baseMap, overrideMap map[string]any

	if err := json.Unmarshal([]byte(base), &baseMap); err != nil {
		return "", fmt.Errorf("base JSON: %w", err)
	}
	if err := json.Unmarshal([]byte(override), &overrideMap); err != nil {
		return "", fmt.Errorf("override JSON: %w", err)
	}

	merged := deepMerge(baseMap, overrideMap)
	out, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func deepMerge(base, override map[string]any) map[string]any {
	result := make(map[string]any, len(base))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		if baseVal, ok := result[k]; ok {
			baseMap, baseIsMap := baseVal.(map[string]any)
			overrideMap, overrideIsMap := v.(map[string]any)
			if baseIsMap && overrideIsMap {
				result[k] = deepMerge(baseMap, overrideMap)
				continue
			}
		}
		result[k] = v
	}
	return result
}
