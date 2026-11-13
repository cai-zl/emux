package tmux

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeTmux writes a shell script that stands in for the tmux binary and
// returns its directory (for PATH injection) and full path.
func fakeTmux(t *testing.T, body string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "tmux")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir, path
}

func TestRunPassesArgsIntact(t *testing.T) {
	_, path := fakeTmux(t, `printf '%s ' "$@"`)
	r := &Runner{TmuxPath: path}
	var out bytes.Buffer
	r.Stdout = &out
	if err := r.Run("split-window", "-h", "-t", "foo"); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "split-window -h -t foo" {
		t.Fatalf("args mangled: %q", got)
	}
}

func TestRunPropagatesExitCode(t *testing.T) {
	_, path := fakeTmux(t, `exit 7`)
	r := &Runner{TmuxPath: path}
	err := r.Run("ls")
	re, ok := err.(*RunError)
	if !ok {
		t.Fatalf("want *RunError, got %T (%v)", err, err)
	}
	if re.ExitCode() != 7 {
		t.Fatalf("exit code = %d, want 7", re.ExitCode())
	}
}

func TestOutputCapturesStdoutAndStderr(t *testing.T) {
	dir, _ := fakeTmux(t, `echo hello; echo boom >&2; exit 2`)
	r := &Runner{TmuxPath: filepath.Join(dir, "tmux")}
	out, err := r.Output("list-sessions")
	if out != "" {
		t.Fatalf("stdout on failure = %q", out)
	}
	re, ok := err.(*RunError)
	if !ok {
		t.Fatalf("want *RunError, got %T (%v)", err, err)
	}
	if !strings.Contains(re.Stderr, "boom") {
		t.Fatalf("stderr not captured: %q", re.Stderr)
	}
	if re.ExitCode() != 2 {
		t.Fatalf("exit code = %d, want 2", re.ExitCode())
	}
}

func TestOutputReturnsTrimmedStdout(t *testing.T) {
	_, path := fakeTmux(t, `printf 'line1\nline2\n\n'`)
	r := &Runner{TmuxPath: path}
	out, err := r.Output("anything")
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	if out != "line1\nline2" {
		t.Fatalf("got %q", out)
	}
}

func TestListSessionsParses(t *testing.T) {
	_, path := fakeTmux(t, `printf 'work\t3\t1700000000\t1\nscratch\t1\t1700000500\t0\n'`)
	r := &Runner{TmuxPath: path}
	ss, err := r.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(ss) != 2 {
		t.Fatalf("got %d sessions, want 2", len(ss))
	}
	if ss[0].Name != "work" || ss[0].Windows != 3 || ss[0].Created != 1700000000 || !ss[0].Attached {
		t.Fatalf("session[0] wrong: %+v", ss[0])
	}
	if ss[1].Name != "scratch" || ss[1].Attached {
		t.Fatalf("session[1] wrong: %+v", ss[1])
	}
}

func TestListSessionsNoServerIsEmpty(t *testing.T) {
	_, path := fakeTmux(t, `echo "no server running on /tmp/x" >&2; exit 1`)
	r := &Runner{TmuxPath: path}
	ss, err := r.ListSessions()
	if err != nil {
		t.Fatalf("no-server should not be an error, got %v", err)
	}
	if len(ss) != 0 {
		t.Fatalf("want empty, got %+v", ss)
	}
}

func TestNewUsesLookPath(t *testing.T) {
	dir, _ := fakeTmux(t, `exit 0`)
	t.Setenv("PATH", dir)
	r, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if r.TmuxPath != filepath.Join(dir, "tmux") {
		t.Fatalf("TmuxPath = %q", r.TmuxPath)
	}

	t.Setenv("PATH", t.TempDir()) // empty dir, no tmux
	_, err = New()
	if err == nil || !strings.Contains(err.Error(), "tmux not found in $PATH") {
		t.Fatalf("want friendly not-found error, got %v", err)
	}
}
