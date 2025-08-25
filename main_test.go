package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const probeTmpl = `
CMAKE={{- range .CMakeDirs}}{{.}}|{{end}}
CPP={{- range .CppExeCandidates}}{{.}}|{{end}}
MAKE={{- range .MakeExeCandidates}}{{.}}|{{end}}
NODE={{- range .NodeEntries}}{{.}}|{{end}}
GOEXE={{- range .GoExeCandidates}}{{.}}|{{end}}
ATTACH={{.AttachOnDev}}
CARGO={{.CargoBinGuess}}
`

func TestTemplateFor_All(t *testing.T) {
	if templateFor(tRust) == "" || templateFor(tGo) == "" || templateFor(tCpp) == "" || templateFor(tNode) == "" {
		t.Fatal("missing embedded template")
	}
}

func TestDetectType(t *testing.T) {
	td := t.TempDir()

	// node
	os.WriteFile(filepath.Join(td, "package.json"), []byte("{}"), 0o644)
	if got, _ := detectType(td); got != tNode {
		t.Fatalf("node: got %s", got)
	}
	os.Remove(filepath.Join(td, "package.json"))

	// cpp (CMake)
	os.WriteFile(filepath.Join(td, "CMakeLists.txt"), []byte(""), 0o644)
	if got, _ := detectType(td); got != tCpp {
		t.Fatalf("cpp/cmake: got %s", got)
	}
	os.Remove(filepath.Join(td, "CMakeLists.txt"))

	// cpp (Make)
	os.WriteFile(filepath.Join(td, "Makefile"), []byte(""), 0o644)
	if got, _ := detectType(td); got != tCpp {
		t.Fatalf("cpp/make: got %s", got)
	}
	os.Remove(filepath.Join(td, "Makefile"))

	// go
	os.WriteFile(filepath.Join(td, "go.mod"), []byte("module x\n"), 0o644)
	if got, _ := detectType(td); got != tGo {
		t.Fatalf("go: got %s", got)
	}
	os.Remove(filepath.Join(td, "go.mod"))
	os.WriteFile(filepath.Join(td, "main.go"), []byte("package main"), 0o644)
	if got, _ := detectType(td); got != tGo {
		t.Fatalf("go by *.go: got %s", got)
	}
	os.Remove(filepath.Join(td, "main.go"))

	// rust
	os.WriteFile(filepath.Join(td, "Cargo.toml"), []byte(""), 0o644)
	if got, _ := detectType(td); got != tRust {
		t.Fatalf("rust: got %s", got)
	}
}

func TestCLI_PrintTemplate_WithFlags_UsesValues(t *testing.T) {
	t.Parallel()

	td := t.TempDir()
	tmpl := filepath.Join(td, "probe.tmpl")
	if err := os.WriteFile(tmpl, []byte(probeTmpl), 0o644); err != nil {
		t.Fatalf("write tmpl: %v", err)
	}

	cmd := exec.Command("go", "run", ".",
		"--print-template",
		"--template", tmpl,
		"--cmake-dirs", "out,.build",
		"--cpp-exes", "app,main,Debug/app",
		"--make-exes", "a.out,main",
		"--node-entries", "dist/index.js,build/index.js",
		"--go-exes", "foo,bin two",
		"--attach-on-dev=false",
		"--cargo-bin", "crate_name",
	)
	outb, err := cmd.CombinedOutput()
	out := string(outb)
	if err != nil {
		t.Fatalf("go run failed: %v\n--- output ---\n%s", err, out)
	}

	expectContains(t, out, "CMAKE=out|.build|")
	expectContains(t, out, "CPP=app|main|Debug/app|")
	expectContains(t, out, "MAKE=a.out|main|")
	expectContains(t, out, "NODE=dist/index.js|build/index.js|")
	expectContains(t, out, "GOEXE=foo|bin two|")
	expectContains(t, out, "ATTACH=false")
	expectContains(t, out, "CARGO=crate_name")
}

func TestCLI_Version_UsesLdflags(t *testing.T) {
	t.Parallel()

	// discover module path for ldflags symbol
	modcmd := exec.Command("go", "list", "-m", "-f", "{{.Path}}")
	modout, err := modcmd.Output()
	if err != nil {
		t.Fatalf("go list -m failed: %v", err)
	}
	module := strings.TrimSpace(string(modout))
	symbol := module + "/internal/buildinfo.Version"

	cmd := exec.Command("go", "run", "-ldflags", "-X "+symbol+"=9.9.9", ".", "version")
	outb, err := cmd.CombinedOutput()
	out := string(outb)
	if err != nil {
		t.Fatalf("go run version failed: %v\n--- output ---\n%s", err, out)
	}
	expectContains(t, out, "justify 9.9.9")
}

func expectContains(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("expected output to contain %q\n--- output ---\n%s", sub, s)
	}
}
