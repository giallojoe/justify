package render_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/giallojoe/justify/internal/render"
)

// Smoke test: the real default template from repo root.
func TestRender_FullTemplate_Smoke(t *testing.T) {
	path := filepath.Join("..", "..", "Justfile.gotpl")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read template (%s): %v", path, err)
	}

	_, err = render.Render(string(b), render.Options{
		CMakeDirs:         []string{"build", ".build", "cmake-build-debug"},
		CppExeCandidates:  []string{"app", "main", "Debug/app", "Debug/main"},
		MakeExeCandidates: []string{"app", "main", "a.out"},
		NodeEntries:       []string{"dist/index.js", "dist/server/index.js", "build/index.js"},
		AttachOnDev:       true,
		CargoBinGuess:     "app",
		GoExeCandidates:   []string{"app", "main"},
	})
	if err != nil {
		t.Fatalf("full template failed to render: %v", err)
	}
}
