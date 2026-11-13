package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/cai-zl/emux/internal/tmux"
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
		Use:   "emux",
		Short: "a friendlier tmux",
		Long:  groupedHelp(),
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
			// verbatim, including leading flags like `emux -2 attach`.
			return a.tmux.Run(args...)
		},
	}
	root.SetHelpCommand(newHelpCmd(a))
	// Completions were explicitly out of scope for emux.
	root.CompletionOptions.DisableDefaultCmd = true
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

// Execute runs emux and returns the process exit code.
func Execute() int {
	return execute(newRootCmd(defaultApp()))
}

// execute is the testable seam behind Execute.
func execute(root *cobra.Command) int {
	err := root.Execute()
	if err != nil {
		w := root.ErrOrStderr()
		switch e := err.(type) {
		case *tmux.RunError:
			// Run() failures already printed tmux's own stderr live;
			// Output()-based failures (session listing) did not.
			if e.Stderr != "" {
				fmt.Fprintln(w, e.Stderr)
			}
		case *quietExit:
			// Already spoke for itself on stdout/stderr.
		default:
			fmt.Fprintln(w, "emux:", err)
		}
	}
	return exitCodeOf(err)
}
