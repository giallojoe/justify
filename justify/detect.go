package justify

import (
	"os"
)

func DetectAll(root string) ([]Target, error) {
	var out []Target

	if _, err := os.Stat(root + "/Cargo.toml"); err == nil {
		out = append(out, Target{Name: "rust-app", Kind: KindRust, Program: root + "/target/debug/app"})
	}
	if _, err := os.Stat(root + "/go.mod"); err == nil {
		out = append(out, Target{Name: "go-app", Kind: KindGo, Program: root})
	}
	if _, err := os.Stat(root + "/package.json"); err == nil {
		out = append(out, Target{Name: "node-app", Kind: KindNode, Program: root + "/index.js"})
	}
	if _, err := os.Stat(root + "/CMakeLists.txt"); err == nil {
		out = append(out, Target{Name: "cpp-app", Kind: KindCPP, Program: root + "/build/app"})
	}

	return out, nil
}
