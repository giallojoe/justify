package render_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/giallojoe/justify/internal/render"
)

const minimalTemplate = `
build:
	@echo {{index .CMakeDirs 0}}
program:
	@echo {{.CargoBinGuess}}
# node: {{range .NodeEntries}}{{.}};{{end}}
attach: {{.AttachOnDev}}
`

func TestRender_MinimalTemplate(t *testing.T) {
	opts := render.Options{
		CMakeDirs:         []string{"build", ".build"},
		CppExeCandidates:  []string{"app", "main"},
		MakeExeCandidates: []string{"a.out"},
		NodeEntries:       []string{"dist/index.js", "build/index.js"},
		AttachOnDev:       true,
		CargoBinGuess:     "mybin",
	}
	out, err := render.Render(minimalTemplate, opts)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	mustContain(t, out, "build:")
	mustContain(t, out, "@echo build")
	mustContain(t, out, "@echo mybin")
	mustContain(t, out, "node: dist/index.js;build/index.js;")
	mustContain(t, out, "attach: true")
	if strings.Contains(out, "\r\n") {
		t.Fatalf("expected normalized LF newlines, found CRLF")
	}
}

func TestRender_CustomTemplate_Golden(t *testing.T) {
	td := t.TempDir()
	// Load template + golden from testdata (so you can tweak examples easily)
	tmplPath := filepath.Join("..", "..", "testdata", "custom_min.gotpl")
	goldPath := filepath.Join("..", "..", "testdata", "custom_min.golden")

	tmplBytes, err := os.ReadFile(tmplPath)
	if err != nil {
		t.Fatalf("read tmpl: %v", err)
	}
	goldBytes, err := os.ReadFile(goldPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	opts := render.Options{
		CMakeDirs:         []string{"cmake-out"},
		CppExeCandidates:  []string{"app"},
		MakeExeCandidates: []string{"main"},
		NodeEntries:       []string{"dist/server/entry.mjs"},
		AttachOnDev:       false,
		CargoBinGuess:     "crate_name",
	}
	got, err := render.Render(string(tmplBytes), opts)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	// Optional: write a copy to temp dir for debugging
	_ = os.WriteFile(filepath.Join(td, "got.just"), []byte(got), 0o644)

	want := string(goldBytes)
	if got != want {
		t.Fatalf("mismatch (-got +want):\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSplitClean(t *testing.T) {
	in := "  a, b ,, c  , "
	out := render.SplitClean(in)
	want := []string{"a", "b", "c"}
	if len(out) != len(want) {
		t.Fatalf("len mismatch: %v vs %v", out, want)
	}
	for i := range want {
		if out[i] != want[i] {
			t.Fatalf("index %d: got %q want %q", i, out[i], want[i])
		}
	}
}

func mustContain(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("expected output to contain %q\n--- output ---\n%s", sub, s)
	}
}
