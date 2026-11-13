package cmd

import (
	"strings"
	"testing"

	"gmux/internal/tmux"
)

// fixture tmux that serves session TSV for list-sessions and echoes its
// args to stdout for anything else.
const sessionScript = `if [ "$1" = "list-sessions" ]; then
printf 'work\t3\t1700000000\t1\nscratch\t1\t1700000500\t0\n'
else
printf '%s ' "$@"
fi
`

func newSessionRunner(t *testing.T, body string) (*tmux.Runner, *strings.Builder) {
	t.Helper()
	r := fakeTmux(t, body)
	var got strings.Builder
	r.Stdout = &got
	return r, &got
}

func TestLsRendersTable(t *testing.T) {
	r, _ := newSessionRunner(t, sessionScript)
	out, err := runRoot(t, r, "ls")
	if err != nil {
		t.Fatalf("ls errored: %v", err)
	}
	// The fixture prints the TSV only when invoked with list-sessions, so
	// seeing the rendered table proves the call happened.
	for _, want := range []string{"NAME", "work", "scratch", "3", "yes", "no"} {
		if !strings.Contains(out, want) {
			t.Fatalf("ls output missing %q:\n%s", want, out)
		}
	}
}

func TestLsNoSessionsFriendly(t *testing.T) {
	r, _ := newSessionRunner(t, `echo "no server running on /tmp/x" >&2; exit 1`)
	out, err := runRoot(t, r, "ls")
	if err != nil {
		t.Fatalf("empty server should exit 0, got %v", err)
	}
	if !strings.Contains(out, "no active sessions") {
		t.Fatalf("want friendly empty message, got %q", out)
	}
}

func TestNewPassesSessionName(t *testing.T) {
	r, tmuxArgs := newSessionRunner(t, `printf '%s ' "$@"`)
	if _, err := runRoot(t, r, "new", "work"); err != nil {
		t.Fatalf("new errored: %v", err)
	}
	if want := "new-session -s work "; tmuxArgs.String() != want {
		t.Fatalf("tmux got %q, want %q", tmuxArgs.String(), want)
	}
}

func TestNewUnnamedAndFlags(t *testing.T) {
	r, tmuxArgs := newSessionRunner(t, `printf '%s ' "$@"`)
	if _, err := runRoot(t, r, "new", "-d", "--dir", "/tmp/proj"); err != nil {
		t.Fatalf("new errored: %v", err)
	}
	if want := "new-session -d -c /tmp/proj "; tmuxArgs.String() != want {
		t.Fatalf("tmux got %q, want %q", tmuxArgs.String(), want)
	}
}

func TestAttachNamed(t *testing.T) {
	r, tmuxArgs := newSessionRunner(t, sessionScript)
	if _, err := runRoot(t, r, "attach", "work"); err != nil {
		t.Fatalf("attach errored: %v", err)
	}
	if want := "attach-session -t work "; tmuxArgs.String() != want {
		t.Fatalf("tmux got %q, want %q", tmuxArgs.String(), want)
	}
}

func TestAttachSingleSessionAttachesDirectly(t *testing.T) {
	single := `if [ "$1" = "list-sessions" ]; then
printf 'solo\t1\t1700000000\t0\n'
else
printf '%s ' "$@"
fi
`
	r, tmuxArgs := newSessionRunner(t, single)
	if _, err := runRoot(t, r, "attach"); err != nil {
		t.Fatalf("attach errored: %v", err)
	}
	if want := "attach-session -t solo "; tmuxArgs.String() != want {
		t.Fatalf("tmux got %q, want %q", tmuxArgs.String(), want)
	}
}

func TestAttachNoSessionsHints(t *testing.T) {
	r, _ := newSessionRunner(t, `echo "no server running on /tmp/x" >&2; exit 1`)
	out, err := runRoot(t, r, "attach")
	if exitCodeOf(err) != 1 {
		t.Fatalf("want exit 1, got %v", err)
	}
	if !strings.Contains(out, "gmux new") {
		t.Fatalf("want hint to run gmux new, got %q", out)
	}
}

func TestAttachMultipleListsAndExits(t *testing.T) {
	r, tmuxArgs := newSessionRunner(t, sessionScript)
	out, err := runRoot(t, r, "attach")
	if exitCodeOf(err) != 1 {
		t.Fatalf("want exit 1, got %v", err)
	}
	if !strings.Contains(out, "work") || !strings.Contains(out, "scratch") {
		t.Fatalf("multi-attach should list sessions, got %q", out)
	}
	if !strings.Contains(out, "gmux attach") {
		t.Fatalf("want retry hint, got %q", out)
	}
	if strings.Contains(tmuxArgs.String(), "attach-session") {
		t.Fatalf("must not attach when ambiguous, tmux args: %q", tmuxArgs.String())
	}
}

func killScript(t *testing.T) (*tmux.Runner, *strings.Builder) {
	t.Helper()
	body := `if [ "$1" = "list-sessions" ]; then
printf 'work\t3\t1700000000\t1\n'
elif [ "$1" = "kill-session" ]; then
printf '%s ' "$@"
else
printf '%s ' "$@"
fi
`
	return newSessionRunner(t, body)
}

func TestKillWithYesFlagSkipsPrompt(t *testing.T) {
	r, tmuxArgs := killScript(t)
	out, err := runRootIn(t, &app{tmux: r}, strings.NewReader(""), "kill", "work", "-y")
	if err != nil {
		t.Fatalf("kill -y errored: %v", err)
	}
	if want := "kill-session -t work "; tmuxArgs.String() != want {
		t.Fatalf("tmux got %q, want %q", tmuxArgs.String(), want)
	}
	if !strings.Contains(out, "work") {
		t.Fatalf("kill should report what it did, got %q", out)
	}
}

func TestKillConfirmProceedsOnY(t *testing.T) {
	r, tmuxArgs := killScript(t)
	_, err := runRootIn(t, &app{tmux: r}, strings.NewReader("y\n"), "kill", "work")
	if err != nil {
		t.Fatalf("confirmed kill errored: %v", err)
	}
	if !strings.Contains(tmuxArgs.String(), "kill-session") {
		t.Fatalf("confirmed kill did not run, tmux args: %q", tmuxArgs.String())
	}
}

func TestKillConfirmAbortsOnN(t *testing.T) {
	r, tmuxArgs := killScript(t)
	out, err := runRootIn(t, &app{tmux: r}, strings.NewReader("n\n"), "kill", "work")
	if exitCodeOf(err) != 1 {
		t.Fatalf("aborted kill should exit 1, got %v", err)
	}
	if strings.Contains(tmuxArgs.String(), "kill-session") {
		t.Fatalf("aborted kill still ran, tmux args: %q", tmuxArgs.String())
	}
	if !strings.Contains(out, "aborted") {
		t.Fatalf("want abort notice, got %q", out)
	}
}

func TestKillNoArgListsSessions(t *testing.T) {
	r, tmuxArgs := killScript(t)
	out, err := runRoot(t, r, "kill")
	if err != nil {
		t.Fatalf("kill without arg errored: %v", err)
	}
	if !strings.Contains(out, "work") {
		t.Fatalf("kill without arg should list sessions, got %q", out)
	}
	if strings.Contains(tmuxArgs.String(), "kill-session") {
		t.Fatalf("kill without arg must not kill anything, tmux args: %q", tmuxArgs.String())
	}
}

func TestRenameOldNew(t *testing.T) {
	r, tmuxArgs := newSessionRunner(t, `printf '%s ' "$@"`)
	if _, err := runRoot(t, r, "rename", "old", "brand-new"); err != nil {
		t.Fatalf("rename errored: %v", err)
	}
	if want := "rename-session -t old brand-new "; tmuxArgs.String() != want {
		t.Fatalf("tmux got %q, want %q", tmuxArgs.String(), want)
	}
}

func TestRenameRequiresTwoArgs(t *testing.T) {
	r, _ := newSessionRunner(t, `printf '%s ' "$@"`)
	for _, args := range [][]string{{"rename"}, {"rename", "only-one"}} {
		if _, err := runRoot(t, r, args...); exitCodeOf(err) != 1 {
			t.Fatalf("rename %v should fail usage, got %v", args, err)
		}
	}
}

func TestDetachInsideTmuxDetachesCurrent(t *testing.T) {
	r, tmuxArgs := newSessionRunner(t, `printf '%s ' "$@"`)
	t.Setenv("TMUX", "/tmp/tmux-0/default,1,0")
	if _, err := runRoot(t, r, "detach"); err != nil {
		t.Fatalf("detach errored: %v", err)
	}
	if want := "detach-client "; tmuxArgs.String() != want {
		t.Fatalf("tmux got %q, want %q", tmuxArgs.String(), want)
	}
}

func TestDetachOutsideWithSessionName(t *testing.T) {
	t.Setenv("TMUX", "")
	r, tmuxArgs := newSessionRunner(t, `printf '%s ' "$@"`)
	if _, err := runRoot(t, r, "detach", "work"); err != nil {
		t.Fatalf("detach errored: %v", err)
	}
	if want := "detach-client -s work "; tmuxArgs.String() != want {
		t.Fatalf("tmux got %q, want %q", tmuxArgs.String(), want)
	}
}

func TestDetachOutsideNoArgSingleSession(t *testing.T) {
	t.Setenv("TMUX", "")
	single := `if [ "$1" = "list-sessions" ]; then
printf 'solo\t1\t1700000000\t0\n'
else
printf '%s ' "$@"
fi
`
	r, tmuxArgs := newSessionRunner(t, single)
	if _, err := runRoot(t, r, "detach"); err != nil {
		t.Fatalf("detach errored: %v", err)
	}
	if want := "detach-client -s solo "; tmuxArgs.String() != want {
		t.Fatalf("tmux got %q, want %q", tmuxArgs.String(), want)
	}
}

func TestDetachOutsideNoArgMultipleLists(t *testing.T) {
	t.Setenv("TMUX", "")
	r, tmuxArgs := newSessionRunner(t, sessionScript)
	out, err := runRoot(t, r, "detach")
	if exitCodeOf(err) != 1 {
		t.Fatalf("ambiguous detach should exit 1, got %v", err)
	}
	if !strings.Contains(out, "work") || !strings.Contains(out, "gmux detach") {
		t.Fatalf("want session list + hint, got %q", out)
	}
	if strings.Contains(tmuxArgs.String(), "detach-client") {
		t.Fatalf("ambiguous detach must not run, tmux args: %q", tmuxArgs.String())
	}
}

func TestDetachOutsideNoArgNoSessions(t *testing.T) {
	t.Setenv("TMUX", "")
	r, _ := newSessionRunner(t, `echo "no server running on /tmp/x" >&2; exit 1`)
	out, err := runRoot(t, r, "detach")
	if err != nil {
		t.Fatalf("detach on empty server should exit 0, got %v", err)
	}
	if !strings.Contains(out, "no active sessions") {
		t.Fatalf("want friendly message, got %q", out)
	}
}

func TestVersionCmd(t *testing.T) {
	r, _ := newSessionRunner(t, `printf '%s ' "$@"`)
	out, err := runRoot(t, r, "version")
	if err != nil {
		t.Fatalf("version errored: %v", err)
	}
	if !strings.Contains(out, "gmux version") {
		t.Fatalf("want version string, got %q", out)
	}
}

func TestVersionFlag(t *testing.T) {
	r, tmuxArgs := newSessionRunner(t, `printf '%s ' "$@"`)
	out, err := runRoot(t, r, "--version")
	if err != nil {
		t.Fatalf("--version errored: %v", err)
	}
	if !strings.Contains(out, "gmux version") {
		t.Fatalf("want version string, got %q", out)
	}
	if tmuxArgs.String() != "" {
		t.Fatalf("--version must not call tmux, got %q", tmuxArgs.String())
	}
}

