package cmd

import (
	"strings"
	"testing"
)

func TestRootHelpGroupsCatalogCommands(t *testing.T) {
	// Exit 97: if help ever reaches tmux, this blows up loudly.
	r, tmuxArgs := newSessionRunner(t, `exit 97`)
	out, err := runRoot(t, r, "--help")
	if err != nil {
		t.Fatalf("--help errored: %v", err)
	}
	if tmuxArgs.String() != "" {
		t.Fatalf("--help must not call tmux, got %q", tmuxArgs.String())
	}
	for _, want := range []string{
		"Sessions", "Windows", "split-window", "refresh-client",
		"emux new", "pass", // grouped catalog + passthrough note
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("--help output missing %q", want)
		}
	}
}

func TestHelpCatalogEntry(t *testing.T) {
	r, tmuxArgs := newSessionRunner(t, `exit 97`)
	out, err := runRoot(t, r, "help", "split-window")
	if err != nil {
		t.Fatalf("help split-window errored: %v", err)
	}
	if tmuxArgs.String() != "" {
		t.Fatalf("catalog hit must not call tmux, got %q", tmuxArgs.String())
	}
	if !strings.Contains(out, "split-window") || !strings.Contains(out, "pane") {
		t.Fatalf("help should show split-window usage + description, got %q", out)
	}
}

func TestHelpGmuxSubcommand(t *testing.T) {
	r, _ := newSessionRunner(t, `exit 97`)
	out, err := runRoot(t, r, "help", "ls")
	if err != nil {
		t.Fatalf("help ls errored: %v", err)
	}
	if !strings.Contains(out, "list sessions") {
		t.Fatalf("help ls should show the emux subcommand help, got %q", out)
	}
}

func TestHelpUnknownPassesThroughToTmux(t *testing.T) {
	r, tmuxArgs := newSessionRunner(t, `printf '%s ' "$@"`)
	if _, err := runRoot(t, r, "help", "definitely-not-a-command"); err != nil {
		t.Fatalf("help passthrough errored: %v", err)
	}
	if want := "help definitely-not-a-command "; tmuxArgs.String() != want {
		t.Fatalf("tmux got %q, want %q", tmuxArgs.String(), want)
	}
}
