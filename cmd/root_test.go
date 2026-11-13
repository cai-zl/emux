package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"easymux/internal/tmux"
)

func fakeTmux(t *testing.T, body string) *tmux.Runner {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "tmux")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return &tmux.Runner{TmuxPath: path}
}

func runRoot(t *testing.T, r *tmux.Runner, args ...string) (string, error) {
	return runRootIn(t, &app{tmux: r}, strings.NewReader(""), args...)
}

func runRootIn(t *testing.T, a *app, in io.Reader, args ...string) (string, error) {
	t.Helper()
	var out strings.Builder
	root := newRootCmd(a)
	root.SetOut(&out)
	root.SetErr(&out)
	if args == nil {
		// SetArgs(nil) makes cobra fall back to os.Args[1:]; force empty.
		args = []string{}
	}
	root.SetArgs(args)
	if in != nil {
		a.in = in
	}
	err := root.Execute()
	return out.String(), err
}

func TestPassthroughForwardsArgsIntact(t *testing.T) {
	var got strings.Builder
	r := fakeTmux(t, `printf '%s ' "$@"`)
	r.Stdout = &got
	if _, err := runRoot(t, r, "split-window", "-h", "-t", "foo"); err != nil {
		t.Fatalf("passthrough failed: %v", err)
	}
	if want := "split-window -h -t foo "; got.String() != want {
		t.Fatalf("tmux got %q, want %q", got.String(), want)
	}
}

func TestPassthroughPropagatesExitCode(t *testing.T) {
	r := fakeTmux(t, `exit 5`)
	_, err := runRoot(t, r, "has-session")
	re, ok := err.(*tmux.RunError)
	if !ok {
		t.Fatalf("want *RunError, got %T (%v)", err, err)
	}
	if re.ExitCode() != 5 {
		t.Fatalf("exit code = %d, want 5", re.ExitCode())
	}
}

func TestBareInvocationShowsHelp(t *testing.T) {
	r := fakeTmux(t, `echo "tmux must not run" >&2; exit 99`)
	out, err := runRoot(t, r)
	if err != nil {
		t.Fatalf("bare easymux errored: %v", err)
	}
	if !strings.Contains(out, "Usage:") || !strings.Contains(out, "easymux") {
		t.Fatalf("bare easymux did not show help: %q", out)
	}
}

func TestHelpFlagShowsHelp(t *testing.T) {
	r := fakeTmux(t, `exit 42`)
	out, err := runRoot(t, r, "--help")
	if err != nil {
		t.Fatalf("--help errored: %v", err)
	}
	if !strings.Contains(out, "Usage:") {
		t.Fatalf("--help did not show help: %q", out)
	}
}

func TestExecutePrintsListFailureStderr(t *testing.T) {
	// ListSessions failing for a non-"no server running" reason must not
	// die silently: tmux's stderr has to reach the user.
	r := fakeTmux(t, `echo "tmux: malformed server socket" >&2; exit 2`)
	var out, errOut strings.Builder
	a := &app{tmux: r}
	root := newRootCmd(a)
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs([]string{"ls"})
	if code := execute(root); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(errOut.String(), "malformed server socket") {
		t.Fatalf("tmux stderr was swallowed, got %q", errOut.String())
	}
}

func TestExitCodeOf(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{nil, 0},
		{&tmux.RunError{Code: 3}, 3},
		{os.ErrNotExist, 1},
	}
	for _, tc := range cases {
		if got := exitCodeOf(tc.err); got != tc.want {
			t.Fatalf("exitCodeOf(%v) = %d, want %d", tc.err, got, tc.want)
		}
	}
}
