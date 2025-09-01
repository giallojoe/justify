package justify

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type state struct {
	LastUsed string `json:"last_used"`
}

func statePath(root string) string {
	return filepath.Join(root, ".justify", "state.json")
}

func ReadLastUsed(root string) (string, error) {
	b, err := os.ReadFile(statePath(root))
	if err != nil {
		return "", err
	}
	var s state
	if err := json.Unmarshal(b, &s); err != nil {
		return "", err
	}
	return s.LastUsed, nil
}

func WriteLastUsed(root, name string) error {
	if err := os.MkdirAll(filepath.Join(root, ".justify"), 0o755); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(state{LastUsed: name}, "", "  ")
	return os.WriteFile(statePath(root), b, 0o644)
}
