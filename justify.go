package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/giallojoe/justify/internal/buildinfo"
	"github.com/giallojoe/justify/internal/render"
)

//go:embed Justfile.gotpl
var defaultTemplate string

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

	cmakeDirs := fs.String("cmake-dirs", "build,.build,cmake-build-debug", "comma-separated CMake build dirs (first used by -B)")
	cppExe := fs.String("cpp-exes", "app,main,Debug/app,Debug/main", "comma-separated executable names inside CMake dirs")
	makeExe := fs.String("make-exes", "app,main,a.out", "comma-separated root-level Makefile targets/exes")
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

	tmplStr := defaultTemplate
	if *templatePath != "" {
		b, err := os.ReadFile(*templatePath)
		if err != nil {
			fail(fmt.Errorf("reading template: %w", err))
		}
		tmplStr = string(b)
	}

	rendered, err := render.Render(tmplStr, render.Options{
		CMakeDirs:         render.SplitClean(*cmakeDirs),
		CppExeCandidates:  render.SplitClean(*cppExe),
		MakeExeCandidates: render.SplitClean(*makeExe),
		NodeEntries:       render.SplitClean(*nodeEntries),
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

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
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
