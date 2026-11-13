package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"gmux/internal/tmux"
)

// app carries the command's dependencies; swapped in tests.
type app struct {
	tmux *tmux.Runner
	in   io.Reader // terminal input for prompts; nil means os.Stdin
}

func defaultApp() *app {
	r, err := tmux.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return &app{tmux: r}
}

func newRootCmd(a *app) *cobra.Command {
	root := &cobra.Command{
		Use:   "gmux",
		Short: "a friendlier tmux",
		Long:  "gmux wraps tmux: friendlier common commands, better help, everything else passes straight through to tmux.",
		// Raw args in, so tmux flags are never intercepted on the
		// passthrough path.
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		SilenceUsage:       true,
		SilenceErrors:      true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
				return c.Help()
			}
			if args[0] == "--version" {
				printVersion(c.OutOrStdout())
				return nil
			}
			// Anything that is not a registered subcommand goes to tmux
			// verbatim, including leading flags like `gmux -2 attach`.
			return a.tmux.Run(args...)
		},
	}
	root.AddCommand(
		newLsCmd(a),
		newNewCmd(a),
		newAttachCmd(a),
		newKillCmd(a),
		newRenameCmd(a),
		newDetachCmd(a),
		newVersionCmd(a),
	)
	return root
}

// exitCodeOf maps an error returned from Execute to a process exit code.
func exitCodeOf(err error) int {
	if err == nil {
		return 0
	}
	if re, ok := err.(*tmux.RunError); ok {
		return re.ExitCode()
	}
	if qe, ok := err.(*quietExit); ok {
		return qe.code
	}
	return 1
}

// quietExit signals a non-zero exit whose message was already printed by the
// subcommand; Execute must not print it again.
type quietExit struct{ code int }

func (e *quietExit) Error() string { return "" }

// sayf prints a line to w and, for code != 0, returns a quiet exit-code error.
func sayf(w io.Writer, code int, format string, args ...any) error {
	fmt.Fprintf(w, format+"\n", args...)
	if code == 0 {
		return nil
	}
	return &quietExit{code}
}

func (a *app) stdin() io.Reader {
	if a.in == nil {
		return os.Stdin
	}
	return a.in
}

// Execute runs gmux and returns the process exit code.
func Execute() int {
	root := newRootCmd(defaultApp())
	err := root.Execute()
	if err != nil {
		switch err.(type) {
		case *tmux.RunError, *quietExit:
			// Already spoke for itself on stdout/stderr.
		default:
			fmt.Fprintln(os.Stderr, "gmux:", err)
		}
	}
	return exitCodeOf(err)
}

var _ io.Writer = os.Stdout // keep io imported for test-facing stream helpers
