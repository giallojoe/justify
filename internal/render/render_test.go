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
# go-exes: {{range .GoExeCandidates}}{{.}}|{{end}}
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
		GoExeCandidates:   []string{"foo", "bar"},
	}
	out, err := render.Render(minimalTemplate, opts)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	mustContain(t, out, "build:")
	mustContain(t, out, "@echo build")
	mustContain(t, out, "@echo mybin")
	mustContain(t, out, "node: dist/index.js;build/index.js;")
	mustContain(t, out, "go-exes: foo|bar|")
	mustContain(t, out, "attach: true")
	if strings.Contains(out, "\r\n") {
		t.Fatalf("expected normalized LF newlines, found CRLF")
	}
}

const goExeCandidatesTemplate = `
program:
	@bash -eu -o pipefail -c '
for c in {{- range .GoExeCandidates }} "{{ . }}" {{- end }}; do
  echo "$c"
done
'`

func TestRender_GoExeCandidates_RendersLoop(t *testing.T) {
	out, err := render.Render(goExeCandidatesTemplate, render.Options{
		GoExeCandidates:   []string{"bin1", "bin two", "bin3"},
		CMakeDirs:         []string{"build"},
		CppExeCandidates:  []string{"app"},
		MakeExeCandidates: []string{"a.out"},
		NodeEntries:       []string{"dist/index.js"},
		AttachOnDev:       true,
		CargoBinGuess:     "mybin",
	})
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	mustContain(t, out, `for c in "bin1" "bin two" "bin3"; do`)
}

func TestRender_CustomTemplate_Golden(t *testing.T) {
	tmplPath := filepath.Join("..", "..", "testdata", "custom_min.tmpl")
	goldPath := filepath.Join("..", "..", "testdata", "custom_min.golden")

	tmplBytes, err := os.ReadFile(tmplPath)
	if err != nil {
		t.Fatalf("read tmpl: %v", err)
	}
	wantBytes, err := os.ReadFile(goldPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	got, err := render.Render(string(tmplBytes), render.Options{
		CMakeDirs:         []string{"cmake-out"},
		CppExeCandidates:  []string{"app"},
		MakeExeCandidates: []string{"main"},
		NodeEntries:       []string{"dist/server/entry.mjs"},
		AttachOnDev:       false,
		CargoBinGuess:     "crate_name",
		GoExeCandidates:   []string{"foo"},
	})
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	if got != string(wantBytes) {
		t.Fatalf("mismatch (-got +want):\n--- got ---\n%s\n--- want ---\n%s", got, string(wantBytes))
	}
}

func TestRender_GoExeCandidates_Golden(t *testing.T) {
	tmplPath := filepath.Join("..", "..", "testdata", "go_exes.tmpl")
	goldPath := filepath.Join("..", "..", "testdata", "go_exes.golden")

	tmplBytes, err := os.ReadFile(tmplPath)
	if err != nil {
		t.Fatalf("read tmpl: %v", err)
	}
	wantBytes, err := os.ReadFile(goldPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	got, err := render.Render(string(tmplBytes), render.Options{
		GoExeCandidates: []string{"bin1", "bin two", "bin3"},
	})
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if got != string(wantBytes) {
		t.Fatalf("mismatch (-got +want):\n--- got ---\n%s\n--- want ---\n%s", got, string(wantBytes))
	}
}

func mustContain(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("expected output to contain %q\n--- output ---\n%s", sub, s)
	}
}
