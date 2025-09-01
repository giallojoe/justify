package justify

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func GenerateVSCode(root string, ts []Target, force bool) error {
	vdir := filepath.Join(root, ".vscode")
	if err := os.MkdirAll(vdir, 0o755); err != nil {
		return err
	}
	tasks := map[string]any{
		"version": "2.0.0",
		"tasks":   []any{},
	}
	launch := map[string]any{
		"version":        "0.2.0",
		"configurations": []any{},
	}

	for _, t := range ts {
		task := map[string]any{
			"label":   t.Name,
			"type":    "shell",
			"command": t.Program,
		}
		tasks["tasks"] = append(tasks["tasks"].([]any), task)

		launchCfg := map[string]any{
			"name": t.Name,
			"type": t.Kind,
			"request": func() string {
				if t.Request != "" {
					return t.Request
				}
				return "launch"
			}(),
			"program": t.Program,
			"cwd":     t.Cwd,
			"args":    t.Args,
			"env":     t.Env,
			"port":    t.Port,
		}
		launch["configurations"] = append(launch["configurations"].([]any), launchCfg)
	}

	tasksPath := filepath.Join(vdir, "tasks.json")
	launchPath := filepath.Join(vdir, "launch.json")

	tasksBytes, _ := json.MarshalIndent(tasks, "", "  ")
	launchBytes, _ := json.MarshalIndent(launch, "", "  ")

	if err := os.WriteFile(tasksPath, tasksBytes, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(launchPath, launchBytes, 0o644); err != nil {
		return err
	}
	return nil
}
