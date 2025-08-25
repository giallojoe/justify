# justify

`justify` is a tiny cross-platform CLI that generates a **standard `Justfile`**
for any project. It keeps each language’s default build/output
layout (Rust, Go, C/C++, Node) and adds portable tasks:

- 🔨 `just build` — builds using each stack’s defaults  
- 🚀 `just run` — builds then runs the auto-discovered program  
- 🔎 `just program` — prints the path to run
  (or `attach:node` for dev servers)  
- 🏷 `just program-kind` — `rust | go | cpp | node | attach | unknown`  
- 📝 `just write-program-env` — writes `.vscode/.program.env` for editors

The CLI renders a `Justfile` from a **template** (`justfile.tmpl`). A default
template is embedded in the binary; you can also provide your own with `--template`.

---

## Install

### From source

```bash
git clone https://github.com/<your-username>/justify.git
cd justify
go build -o justify .
# optional: put it in your PATH
mv justify ~/.local/bin/    # Linux/macOS
# or on Windows
Move-Item .\justify.exe $env:USERPROFILE\bin\
```

### Prebuilt binaries

Download from **Releases** and put the binary in your PATH:

- `justify-linux-amd64`, `justify-linux-arm64`
- `justify-darwin-amd64`, `justify-darwin-arm64`
- `justify-windows-amd64.exe`

---

## Quick start

In any project folder:

```bash
justify              # writes a Justfile using the embedded template
just build
just program         # prints the path to execute (or "attach:node")
just run             # builds + runs
```

Show help & manual (embedded):

```bash
justify --help
justify manual
justify version
```

Render without writing:

```bash
justify --print-template
```

Use a custom template:

```bash
justify --template ./justfile.tmpl --force -o ./tools/Justfile
```

---

## Template options

The template is rendered with these fields (with sensible defaults):

- `CMakeDirs` — list of build dirs to probe (first used by `cmake -B`)  
- `CppExeCandidates` — candidate executable names inside build dirs  
- `MakeExeCandidates` — root-level executables for Make builds  
- `NodeEntries` — common built Node entry files to try  
- `AttachOnDev` — if `true`, `program` prints `attach:node` when only
  a dev server exists  
- `CargoBinGuess` — fallback Cargo bin name (defaults to folder name)

You can override via flags:

```bash
justify \
  --cmake-dirs "out/build/debug,.build" \
  --cpp-exes "myapp,main" \
  --make-exes "app,main,a.out" \
  --node-entries "dist/server/entry.mjs,.next/standalone/server.js" \
  --attach-on-dev=false \
  --cargo-bin "my_crate"
```

---

## Editor integration (VS Code)

```jsonc
// .vscode/tasks.json
{
  "version": "2.0.0",
  "tasks": [
    { "label": "Build", "type": "shell", "command": "just build" },
    { 
      "label": "Resolve Program",
      "type": "shell",
      "command": "just write-program-env",
      "dependsOn": "Build" 
    }
  ]
}
```

```jsonc
// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Native (auto)",
      "type": "cppdbg", // or "lldb" on macOS
      "request": "launch",
      "program": "${env:PROGRAM}",
      "cwd": "${workspaceFolder}",
      "preLaunchTask": "Resolve Program",
      "envFile": "${workspaceFolder}/.vscode/.program.env"
    },
    {
      "name": "Node (auto)",
      "type": "node",
      "request": "launch",
      "program": "${env:PROGRAM}",
      "cwd": "${workspaceFolder}",
      "preLaunchTask": "Resolve Program"
    },
    {
      "name": "Node: Attach (auto)",
      "type": "node",
      "request": "attach",
      "port": 9229,
      "preLaunchTask": "Resolve Program"
    }
  ]
}
```

## Contributing

- Ensure Go ≥ 1.22
- Keep `justfile.tmpl` POSIX-shell friendly
- Run `go test ./...` before PRs
