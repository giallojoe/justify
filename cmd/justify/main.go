package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/giallojoe/justify/justify"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "help", "-h", "--help":
		usage()
	case "version", "-v", "--version":
		fmt.Println("justify", justify.Version)
	case "targets":
		cmdTargets(os.Args[2:])
	case "program":
		cmdProgram(os.Args[2:])
	case "target":
		cmdTarget(os.Args[2:])
	case "editor":
		cmdEditor(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Print(`justify — editor scaffolding for debug/run

Commands:
  targets [--json] [--init] [--force]
  program [--target NAME]
  target set NAME
  target add --name NAME --kind KIND --program PATH [--port N] [--request launch|attach] [--cwd DIR] [--arg ...] [--env K=V ...]
  editor --editor vscode|neovim|neovim-global|both [--force]
`)
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

func cmdTargets(args []string) {
	fs := flag.NewFlagSet("targets", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "")
	init := fs.Bool("init", false, "")
	force := fs.Bool("force", false, "")
	_ = fs.Parse(args)

	if *init {
		ts, err := justify.DetectAll(".")
		if err != nil {
			die(err)
		}
		f := justify.File{Version: 1, Targets: ts}
		if err := justify.SaveTargets(".", f, *force); err != nil {
			die(err)
		}
		fmt.Println("Wrote .justify/targets.json")
		return
	}

	ts, err := justify.List(".")
	if err != nil {
		die(err)
	}

	if *jsonOut {
		b, _ := json.MarshalIndent(ts, "", "  ")
		fmt.Println(string(b))
		return
	}
	for _, t := range ts {
		fmt.Printf("%-16s %-5s %s\n", t.Name, t.Kind, t.Program)
	}
}

func cmdProgram(args []string) {
	fs := flag.NewFlagSet("program", flag.ExitOnError)
	name := fs.String("target", "", "")
	_ = fs.Parse(args)
	p, err := justify.ResolveProgram(".", *name)
	if err != nil {
		die(err)
	}
	fmt.Println(p)
}

func cmdTarget(args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(2)
	}
	switch args[0] {
	case "set":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "usage: justify target set NAME")
			os.Exit(2)
		}
		if err := justify.WriteLastUsed(".", args[1]); err != nil {
			die(err)
		}
	case "add":
		// TODO: parse flags for --name, --kind, etc.
		fmt.Fprintln(os.Stderr, "not implemented yet: justify target add")
		os.Exit(2)
	default:
		fmt.Fprintln(os.Stderr, "usage: justify target {set|add} …")
		os.Exit(2)
	}
}

func cmdEditor(args []string) {
	fs := flag.NewFlagSet("editor", flag.ExitOnError)
	editor := fs.String("editor", "vscode", "")
	force := fs.Bool("force", false, "")
	out := fs.String("out", "", "output path (for neovim-global); use '-' for stdout")
	_ = fs.Parse(args)

	// For neovim-global we don't need targets; for others we do
	switch *editor {
	case "neovim-global":
		outPath := *out
		if outPath == "" {
			if home, _ := os.UserHomeDir(); home != "" {
				outPath = filepath.Join(home, ".config", "nvim", "lua", "justify_client.lua")
			} else {
				outPath = "-" // fallback to stdout
			}
		} else if outPath == "-" {
			outPath = "-"
		}
		if err := justify.WriteNeovimClient(outPath, *force); err != nil {
			die(err)
		}
		if outPath != "-" {
			fmt.Println("Wrote", outPath)
		}
		return
	}

	// otherwise, generate project-local files using detected/declared targets
	ts, err := justify.List(".")
	if err != nil {
		die(err)
	}
	switch *editor {
	case "vscode":
		if err := justify.GenerateVSCode(".", ts, *force); err != nil {
			die(err)
		}
		fmt.Println("Wrote .vscode/tasks.json and .vscode/launch.json")
	case "neovim":
		if err := justify.GenerateNeovim(".", ts, *force); err != nil {
			die(err)
		}
		fmt.Println("Wrote .justify/dap.lua")
	case "both":
		if err := justify.GenerateVSCode(".", ts, *force); err != nil {
			die(err)
		}
		if err := justify.GenerateNeovim(".", ts, *force); err != nil {
			die(err)
		}
		fmt.Println("Wrote VS Code + Neovim configs")
	default:
		fmt.Fprintln(os.Stderr, "unknown --editor:", *editor)
		os.Exit(2)
	}
}
