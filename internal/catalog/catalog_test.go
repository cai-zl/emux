package catalog

import "testing"

func TestMustLoadIntegrity(t *testing.T) {
	cmds := MustLoad()
	if len(cmds) < 80 {
		t.Fatalf("expected at least 80 tmux commands, got %d", len(cmds))
	}

	validCat := map[string]bool{}
	for _, c := range CategoryOrder {
		validCat[c] = true
	}

	seen := map[string]bool{}
	aliasOwners := map[string]string{}
	for _, e := range cmds {
		if e.Name == "" {
			t.Fatalf("entry with empty name: %+v", e)
		}
		if e.Usage == "" {
			t.Fatalf("entry %s: empty usage", e.Name)
		}
		if e.Description == "" {
			t.Fatalf("entry %s: empty description", e.Name)
		}
		if !validCat[e.Category] {
			t.Fatalf("entry %s: unknown category %q", e.Name, e.Category)
		}
		if seen[e.Name] {
			t.Fatalf("duplicate command %s", e.Name)
		}
		seen[e.Name] = true
		for _, a := range e.Aliases {
			if prev, dup := aliasOwners[a]; dup {
				t.Fatalf("alias %q claimed by both %s and %s", a, prev, e.Name)
			}
			aliasOwners[a] = e.Name
		}
	}
}

func TestMustLoadSpotChecks(t *testing.T) {
	cases := []struct {
		query    string
		wantName string
		alias    string
		category string
	}{
		{"new-session", "new-session", "new", "Sessions"},
		{"attach", "attach-session", "attach", "Sessions"},
		{"ls", "list-sessions", "ls", "Sessions"},
		{"splitw", "split-window", "splitw", "Panes"},
		{"bind-key", "bind-key", "bind", "Key Bindings & Menus"},
		{"set", "set-option", "set", "Options & Environment"},
	}
	for _, tc := range cases {
		e, ok := Lookup(tc.query)
		if !ok {
			t.Fatalf("Lookup(%q) not found", tc.query)
		}
		if e.Name != tc.wantName {
			t.Fatalf("Lookup(%q) = %s, want %s", tc.query, e.Name, tc.wantName)
		}
		if !hasAlias(e, tc.alias) {
			t.Fatalf("%s aliases %v missing %q", e.Name, e.Aliases, tc.alias)
		}
		if e.Category != tc.category {
			t.Fatalf("%s category = %q, want %q", e.Name, e.Category, tc.category)
		}
	}
}

func TestLookupUnknown(t *testing.T) {
	if _, ok := Lookup("no-such-command-xyz"); ok {
		t.Fatal("Lookup should miss unknown commands")
	}
}

func TestEveryCategoryPopulated(t *testing.T) {
	cmds := MustLoad()
	counts := map[string]int{}
	for _, e := range cmds {
		counts[e.Category]++
	}
	for _, c := range CategoryOrder {
		if counts[c] == 0 {
			t.Fatalf("category %q has no entries", c)
		}
	}
}

func hasAlias(e Entry, alias string) bool {
	for _, a := range e.Aliases {
		if a == alias {
			return true
		}
	}
	return false
}
