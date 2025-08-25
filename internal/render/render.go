package render

import (
	"strings"
	"text/template"
)

type Options struct {
	CMakeDirs         []string
	CppExeCandidates  []string
	MakeExeCandidates []string
	GoExeCandidates   []string
	NodeEntries       []string
	AttachOnDev       bool
	CargoBinGuess     string
}

// Render applies opts to tmplStr and returns the result with normalized \n.
func Render(tmplStr string, opts Options) (string, error) {
	tmpl, err := template.New("justfile").Parse(normalizeLF(tmplStr))
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := tmpl.Execute(&b, opts); err != nil {
		return "", err
	}
	return b.String(), nil
}

func normalizeLF(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }

// Helpers to keep tests focused (you can export or mirror them from main if you want)
func SplitClean(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
