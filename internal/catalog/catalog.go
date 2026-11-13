// Package catalog holds the static, embedded catalog of tmux commands:
// usage syntax, aliases, a short English description, and a category for
// each. It reflects tmux 3.x and is authored by hand, so it may drift
// slightly from a future tmux release.
package catalog

import (
	"fmt"
	"sort"

	_ "embed"

	"gopkg.in/yaml.v3"
)

// CategoryOrder is the display order of categories; anything not listed
// sorts after these.
var CategoryOrder = []string{
	"Sessions",
	"Windows",
	"Panes",
	"Key Bindings & Menus",
	"Copy Mode & Buffers",
	"Options & Environment",
	"Miscellaneous",
}

// Entry describes one tmux command.
type Entry struct {
	Name        string   `yaml:"name"`
	Aliases     []string `yaml:"aliases"`
	Usage       string   `yaml:"usage"`
	Description string   `yaml:"description"`
	Category    string   `yaml:"category"`
}

//go:embed catalog.yaml
var raw []byte

var (
	entries    []Entry
	lookupIdx  map[string]Entry
	loadErr    error
)

func load() ([]Entry, error) {
	var cmds []Entry
	if err := yaml.Unmarshal(raw, &cmds); err != nil {
		return nil, fmt.Errorf("catalog: parse embedded catalog.yaml: %w", err)
	}
	return cmds, nil
}

func index(cmds []Entry) map[string]Entry {
	idx := make(map[string]Entry, len(cmds)*2)
	for _, e := range cmds {
		idx[e.Name] = e
		for _, a := range e.Aliases {
			idx[a] = e
		}
	}
	return idx
}

func init() {
	entries, loadErr = load()
	if loadErr == nil {
		lookupIdx = index(entries)
	}
}

// MustLoad returns the catalog entries, panicking if the embedded YAML is
// invalid. That is a programmer error caught by tests, not a runtime path.
func MustLoad() []Entry {
	if loadErr != nil {
		panic(loadErr)
	}
	return entries
}

// Lookup resolves a command by name or alias.
func Lookup(nameOrAlias string) (Entry, bool) {
	e, ok := lookupIdx[nameOrAlias]
	return e, ok
}

// ByCategory groups entries by category, each group sorted by name, with
// categories returned in display order.
func ByCategory(cmds []Entry) map[string][]Entry {
	groups := make(map[string][]Entry)
	for _, e := range cmds {
		groups[e.Category] = append(groups[e.Category], e)
	}
	for _, g := range groups {
		sort.Slice(g, func(i, j int) bool { return g[i].Name < g[j].Name })
	}
	return groups
}
