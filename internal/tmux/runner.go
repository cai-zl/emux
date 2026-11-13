// Package tmux runs the tmux binary: pass-through execution with stdio
// wiring, output capture for parsing, and session listing.
package tmux

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// RunError is tmux's non-zero exit, carrying captured stderr if any.
type RunError struct {
	Code   int
	Stderr string
}

func (e *RunError) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("tmux exited with status %d: %s", e.Code, strings.TrimSpace(e.Stderr))
	}
	return fmt.Sprintf("tmux exited with status %d", e.Code)
}

func (e *RunError) ExitCode() int { return e.Code }

// Runner executes tmux. Stdout/Stderr/Stdin default to the process streams
// and may be replaced in tests.
type Runner struct {
	TmuxPath string
	Stdout   io.Writer
	Stderr   io.Writer
	Stdin    io.Reader
}

// New resolves tmux on $PATH.
func New() (*Runner, error) {
	path, err := exec.LookPath("tmux")
	if err != nil {
		return nil, fmt.Errorf("tmux not found in $PATH; install tmux first")
	}
	return &Runner{TmuxPath: path}, nil
}

func (r *Runner) stdio() (io.Writer, io.Writer, io.Reader) {
	out, errOut, in := r.Stdout, r.Stderr, r.Stdin
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	if in == nil {
		in = os.Stdin
	}
	return out, errOut, in
}

// Run executes tmux with stdio wired straight through, so interactive
// commands like attach-session work. Non-zero exits return *RunError.
func (r *Runner) Run(args ...string) error {
	out, errOut, in := r.stdio()
	cmd := exec.Command(r.TmuxPath, args...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = out, errOut, in
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return &RunError{Code: ee.ExitCode()}
		}
		return err
	}
	return nil
}

// Output runs tmux capturing stdout (trimmed). Stderr and the exit code
// ride back in *RunError on failure.
func (r *Runner) Output(args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(r.TmuxPath, args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return "", &RunError{Code: ee.ExitCode(), Stderr: stderr.String()}
		}
		return "", err
	}
	return strings.Trim(stdout.String(), " \t\n"), nil
}

// Session is one tmux session.
type Session struct {
	Name     string
	Windows  int
	Created  int64
	Attached bool
}

const sessionFormat = "#{session_name}\t#{session_windows}\t#{session_created}\t#{session_attached}"

// ListSessions returns all sessions; a server that is not running is not
// an error — it simply has no sessions.
func (r *Runner) ListSessions() ([]Session, error) {
	out, err := r.Output("list-sessions", "-F", sessionFormat)
	if err != nil {
		if re, ok := err.(*RunError); ok && strings.Contains(re.Stderr, "no server running") {
			return nil, nil
		}
		return nil, err
	}
	return ParseSessions(out), nil
}

// ParseSessions parses list-sessions TSV output.
func ParseSessions(out string) []Session {
	if out == "" {
		return nil
	}
	var ss []Session
	for _, line := range strings.Split(out, "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) < 4 {
			continue
		}
		windows, _ := strconv.Atoi(parts[1])
		created, _ := strconv.ParseInt(parts[2], 10, 64)
		ss = append(ss, Session{
			Name:     parts[0],
			Windows:  windows,
			Created:  created,
			Attached: parts[3] == "1",
		})
	}
	return ss
}
