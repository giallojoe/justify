package render_test

import (
	"strings"
	"testing"

	"github.com/giallojoe/justify/internal/render"
)

// Probes every field to ensure expansion is wired correctly.
const probeTemplate = `
# RUST
RUST_BIN={{.CargoBinGuess}}

# GO
GO_EXES={{- range .GoExeCandidates}}{{.}}|{{- end}}

# CMAKE
CMAKE_DIRS={{- range .CMakeDirs}}{{.}}|{{- end}}
CMAKE_EXES={{- range .CppExeCandidates}}{{.}}|{{- end}}

# MAKE
MAKE_EXES={{- range .MakeExeCandidates}}{{.}}|{{- end}}

# NODE
NODE_ENTRIES={{- range .NodeEntries}}{{.}}|{{- end}}
ATTACH_ON_DEV={{.AttachOnDev}}
`

func TestRender_AllProjectTypes_Probe(t *testing.T) {
	opts := render.Options{
		CMakeDirs:         []string{"build", ".build", "cmake-build-debug"},
		CppExeCandidates:  []string{"app", "main", "Debug/app", "Debug/main"},
		MakeExeCandidates: []string{"app", "main", "a.out"},
		NodeEntries:       []string{"dist/index.js", "dist/server/index.js", "build/index.js"},
		AttachOnDev:       true,
		CargoBinGuess:     "my_crate",
		GoExeCandidates:   []string{"tool", "cmd-app"},
	}
	out, err := render.Render(probeTemplate, opts)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	mustContain(t, out, "RUST_BIN=my_crate")
	mustContain(t, out, "GO_EXES=tool|cmd-app|")
	mustContain(t, out, "CMAKE_DIRS=build|.build|cmake-build-debug|")
	mustContain(t, out, "CMAKE_EXES=app|main|Debug/app|Debug/main|")
	mustContain(t, out, "MAKE_EXES=app|main|a.out|")
	mustContain(t, out, "NODE_ENTRIES=dist/index.js|dist/server/index.js|build/index.js|")
	mustContain(t, out, "ATTACH_ON_DEV=true")

	if strings.Contains(out, "\r\n") {
		t.Fatalf("expected LF newlines, found CRLF")
	}
}

func TestRender_Node_AttachFlagVariants(t *testing.T) {
	tmpl := `ATTACH_ON_DEV={{.AttachOnDev}}`
	outTrue, err := render.Render(tmpl, render.Options{AttachOnDev: true})
	if err != nil {
		t.Fatal(err)
	}
	outFalse, err := render.Render(tmpl, render.Options{AttachOnDev: false})
	if err != nil {
		t.Fatal(err)
	}

	mustContain(t, outTrue, "ATTACH_ON_DEV=true")
	mustContain(t, outFalse, "ATTACH_ON_DEV=false")
}
