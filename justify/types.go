package justify

type Kind string

const (
	KindRust Kind = "rust"
	KindGo   Kind = "go"
	KindCPP  Kind = "cpp"
	KindNode Kind = "node"
)

type Target struct {
	Name    string            `json:"name"`
	Kind    Kind              `json:"kind"`
	Program string            `json:"program"` // path or "attach:node"
	Cwd     string            `json:"cwd,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Request string            `json:"request,omitempty"` // "launch"|"attach"
	Port    int               `json:"port,omitempty"`
}

type File struct {
	Version int      `json:"version"`
	Targets []Target `json:"targets"`
}
