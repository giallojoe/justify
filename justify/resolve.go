package justify

import (
	"errors"
	"os"
)

func List(root string) ([]Target, error) {
	if _, err := os.Stat(targetsPath(root)); err == nil {
		f, err := LoadTargets(root)
		if err != nil {
			return nil, err
		}
		return f.Targets, nil
	}
	return DetectAll(root)
}

func FindByName(ts []Target, name string) *Target {
	for i := range ts {
		if ts[i].Name == name {
			return &ts[i]
		}
	}
	return nil
}

func ResolveProgram(root, name string) (string, error) {
	ts, err := List(root)
	if err != nil {
		return "", err
	}
	if name != "" {
		if t := FindByName(ts, name); t != nil {
			return t.Program, nil
		}
		return "", errors.New("unknown target: " + name)
	}
	if lu, _ := ReadLastUsed(root); lu != "" {
		if t := FindByName(ts, lu); t != nil {
			return t.Program, nil
		}
	}
	if len(ts) == 0 {
		return "", errors.New("no targets")
	}
	return ts[0].Program, nil
}
