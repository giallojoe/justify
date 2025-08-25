package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/giallojoe/justify/internal/buildinfo"
	"github.com/giallojoe/justify/internal/render"
)

//go:embed templates/Justfile.rust.gotpl
var tplRust string

//go:embed templates/Justfile.go.gotpl
var tplGo string

//go:embed templates/Justfile.cpp.gotpl
var tplCpp string

//go:embed templates/Justfile.node.gotpl
var tplNode string

type projType string

const (
	tAuto projType = "auto"
	tRust projType = "rust"
	tGo   projType = "go"
	tCpp  projType = "cpp"
	tNode projType = "node"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help", "-", "--help":
			fmt.Print(helpText())
			return
		case "manual":
			fmt.Print(manualText())
			return
		case "version", "-v", "--version":
			fmt.Println("justify", buildinfo.Version)
			return
		}
	}

	fs := flag.NewFlagSet("justify", flag.ContinueOnError)
	fs.SetOutput(new(noopWriter))
	out := fs.String("o", "Justfile", "output filename")
	force := fs.Bool("force", false, "overwrite output if exists")
	printTemplate := fs.Bool("print-template", false, "render the template to stdout")
	templatePath := fs.String("template", "", "path to a custom template (optional)")
	typ := fs.String("type", "auto", "project type: auto|rust|go|cpp|node")

	cmakeDirs := fs.String("cmake-dirs", "build,.build,cmake-build-debug", "comma-separated CMake build dirs (first used by -B)")
	cppExe := fs.String("cpp-exes", "app,main,Debug/app,Debug/main", "comma-separated executable names inside CMake dirs")
	makeExe := fs.String("make-exes", "app,main,a.out", "comma-separated root-level Makefile targets/exes")
	goExes := fs.String("go-exes", "app,main", "comma-separated Go binary names to try")
	nodeEntries := fs.String("node-entries", "dist/index.js,dist/server/index.js,dist/server/entry.mjs,build/index.js,.next/standalone/server.js", "comma-separated Node build entry files")
	attachOnDev := fs.Bool("attach-on-dev", true, "emit attach:node when only dev server is present")
	cargoBinGuess := fs.String("cargo-bin", guessFolderName(), "fallback cargo bin name (default: folder name)")

	if err := fs.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Print(helpText())
			return
		}
		fail(err)
	}

	// choose template
	var tmpl string
	if *templatePath != "" {
		b, err := os.ReadFile(*templatePath)
		if err != nil {
			fail(fmt.Errorf("read template: %w", err))
		}
		tmpl = string(b)
	} else {
		t := projType(*typ)
		if t == tAuto {
			d, err := detectType(".")
			if err != nil {
				fail(err)
			}
			t = d
			fmt.Fprintf(os.Stderr, "Detected project type: %s\n", t)
		}
		tmpl = templateFor(t)
		if tmpl == "" {
			fail(fmt.Errorf("no template for type %q", *typ))
		}
	}

	rendered, err := render.Render(tmpl, render.Options{
		CMakeDirs:         render.SplitClean(*cmakeDirs),
		CppExeCandidates:  render.SplitClean(*cppExe),
		MakeExeCandidates: render.SplitClean(*makeExe),
		NodeEntries:       render.SplitClean(*nodeEntries),
		GoExeCandidates:   render.SplitClean(*goExes),
		AttachOnDev:       *attachOnDev,
		CargoBinGuess:     *cargoBinGuess,
	})
	if err != nil {
		fail(fmt.Errorf("render: %w", err))
	}

	if *printTemplate {
		fmt.Print(rendered)
		return
	}

	dst := *out
	if exists(dst) && !*force {
		fmt.Fprintf(os.Stderr, "Refusing to overwrite %s (use --force to override)\n", dst)
		os.Exit(1)
	}

	if err := os.WriteFile(dst, []byte(rendered), 0o644); err != nil {
		fail(err)
	}
	fmt.Printf("✅ Wrote %s\n", dst)
}

func helpText() string {
	return `justify — generate a standard Justfile from a template

USAGE
  justify [flags]
  justify help
  justify manual
  justify version

FLAGS
  -o <file>            Output filename (default: Justfile)
  --force              Overwrite the output file if it already exists
  --template <path>    Use a custom template instead of the embedded one
  --print-template     Print the rendered template to stdout

  --cmake-dirs "build,.build,cmake-build-debug"
  --cpp-exes "app,main,Debug/app,Debug/main"
  --make-exes "app,main,a.out"
  --node-entries "dist/index.js,dist/server/index.js,dist/server/entry.mjs,build/index.js,.next/standalone/server.js"
  --attach-on-dev (true|false)
  --cargo-bin "<name>" (defaults to folder name)

EXAMPLES
  justify
  justify --force -o .vscode/Justfile
  justify --node-entries "dist/server/entry.mjs,.next/standalone/server.js"
  justify --template ./my-justfile.tmpl --print-template
`
}

func manualText() string {
	return `justify manual

OVERVIEW
  justify renders a Justfile from a Go template. The default template is embedded
  into the binary; you can also provide your own with --template.

TEMPLATE DATA
  .CMakeDirs          []string   list of CMake build directories to probe
  .CppExeCandidates   []string   candidate executable names inside build dirs
  .MakeExeCandidates  []string   root-level executable names for Makefile builds
  .NodeEntries        []string   common Node built entries to try
  .AttachOnDev        bool       if true, program prints 'attach:node' on dev-only
  .CargoBinGuess      string     fallback cargo bin name

WORKFLOW
  1) Run 'justify' to write Justfile
  2) 'just build' builds using defaults
  3) 'just program' returns the executable to run or 'attach:node'

TIPS
  - Use --template to point to a team-specific template
  - Use --print-template for CI to validate rendering
`
}

type noopWriter struct{}

func (noopWriter) Write(p []byte) (int, error) { return len(p), nil }

func detectType(dir string) (projType, error) {
	// priority: rust > go > cpp > node
	if exists(filepath.Join(dir, "Cargo.toml")) {
		return tRust, nil
	}
	if exists(filepath.Join(dir, "go.mod")) || hasExt(dir, ".go") {
		return tGo, nil
	}
	if exists(filepath.Join(dir, "CMakeLists.txt")) || exists(filepath.Join(dir, "Makefile")) {
		return tCpp, nil
	}
	if exists(filepath.Join(dir, "package.json")) {
		return tNode, nil
	}
	return "", fmt.Errorf("unable to detect project type (use --type or --template)")
}

func templateFor(t projType) string {
	switch t {
	case tRust:
		return tplRust
	case tGo:
		return tplGo
	case tCpp:
		return tplCpp
	case tNode:
		return tplNode
	default:
		return ""
	}
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }
func hasExt(dir, ext string) bool {
	ents, _ := os.ReadDir(dir)
	for _, e := range ents {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ext) {
			return true
		}
	}
	return false
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

func guessFolderName() string {
	wd, err := os.Getwd()
	if err != nil {
		return "app"
	}
	return filepath.Base(wd)
}
