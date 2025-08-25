// render_smoke_test.go
package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// buildOnce builds the justify binary into t.TempDir() and returns its path.
func buildOnce(t *testing.T) string {
	t.Helper()
	td := t.TempDir()
	bin := filepath.Join(td, "justify")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, ".")
	// Build from repo root (tests run in module), so no need to set Dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n--- output ---\n%s", err, string(out))
	}
	return bin
}

// runBin executes the compiled binary with given args and working dir, returns stdout+stderr.
func runBin(t *testing.T, bin, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n--- output ---\n%s", filepath.Base(bin), args, err, string(out))
	}
	return string(out)
}

func TestPrintTemplate_ByType(t *testing.T) {
	bin := buildOnce(t)

	cases := []struct {
		typ    string
		expect []string
	}{
		{"rust", []string{"cargo build", "cargo run", "cargo test"}},
		{"go", []string{"go build ./...", "go run .", "go test ./..."}},
		{"cpp", []string{"cmake -S . -B build", "cmake --build build", "make"}},
		{"node", []string{"npm run build", "npm start", "npm run dev", "node ."}},
	}

	for _, tc := range cases {
		t.Run(tc.typ, func(t *testing.T) {
			out := runBin(t, bin, "", "--print-template", "--type", tc.typ)
			for _, want := range tc.expect {
				if !strings.Contains(out, want) {
					t.Fatalf("%s template missing %q\n--- output ---\n%s", tc.typ, want, out)
				}
			}
		})
	}
}

func TestPrintTemplate_AutoDetect(t *testing.T) {
	bin := buildOnce(t)

	type detector struct {
		name   string
		files  map[string]string
		expect string
	}
	tests := []detector{
		{
			name:   "rust",
			files:  map[string]string{"Cargo.toml": "[package]\nname=\"x\"\nversion=\"0.0.0\"\n"},
			expect: "cargo build",
		},
		{
			name:   "go",
			files:  map[string]string{"go.mod": "module example.com/x\n"},
			expect: "go build ./...",
		},
		{
			name:   "cpp-cmake",
			files:  map[string]string{"CMakeLists.txt": "cmake_minimum_required(VERSION 3.22)\nproject(x)\n"},
			expect: "cmake -S . -B build",
		},
		{
			name:   "cpp-make",
			files:  map[string]string{"Makefile": "all:\n\t@echo building\n"},
			expect: "make",
		},
		{
			name:   "node",
			files:  map[string]string{"package.json": "{}\n"},
			expect: "npm run build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := t.TempDir()
			for p, content := range tt.files {
				if err := os.WriteFile(filepath.Join(td, p), []byte(content), 0o644); err != nil {
					t.Fatalf("write %s: %v", p, err)
				}
			}
			out := runBin(t, bin, td, "--print-template") // auto-detect in td
			if !strings.Contains(out, tt.expect) {
				t.Fatalf("auto-detect %s: expected %q in output\n--- output ---\n%s", tt.name, tt.expect, out)
			}
		})
	}
}

func TestWriteJustfile_Force(t *testing.T) {
	bin := buildOnce(t)
	td := t.TempDir()
	justPath := filepath.Join(td, "Justfile")

	// create a node project marker and write the file
	if err := os.WriteFile(filepath.Join(td, "package.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	_ = runBin(t, bin, td) // default: writes Justfile

	// overwrite with --force and custom -o (still Justfile here)
	_ = runBin(t, bin, td, "--force", "-o", "Justfile")

	// sanity check file exists and has expected content
	b, err := os.ReadFile(justPath)
	if err != nil {
		t.Fatalf("read Justfile: %v", err)
	}
	if !strings.Contains(string(b), "npm run build") {
		t.Fatalf("written Justfile does not look like node template\n--- file ---\n%s", string(b))
	}
}
