package justify

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func targetsPath(root string) string {
	return filepath.Join(root, ".justify", "targets.json")
}

func LoadTargets(root string) (File, error) {
	var f File
	b, err := os.ReadFile(targetsPath(root))
	if err != nil {
		return f, err
	}
	if err := json.Unmarshal(b, &f); err != nil {
		return f, err
	}
	if f.Version == 0 {
		f.Version = 1
	}
	return f, nil
}

func SaveTargets(root string, f File, force bool) error {
	if err := os.MkdirAll(filepath.Join(root, ".justify"), 0o755); err != nil {
		return err
	}
	dst := targetsPath(root)
	if !force {
		if _, err := os.Stat(dst); err == nil {
			return errors.New(dst + " exists (use --force)")
		}
	}
	b, _ := json.MarshalIndent(f, "", "  ")
	return os.WriteFile(dst, b, 0o644)
}

func UpsertTarget(root string, t Target) error {
	f, err := LoadTargets(root)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		f = File{Version: 1}
	}
	replaced := false
	for i := range f.Targets {
		if f.Targets[i].Name == t.Name {
			f.Targets[i] = t
			replaced = true
			break
		}
	}
	if !replaced {
		f.Targets = append(f.Targets, t)
	}
	return SaveTargets(root, f, true)
}
