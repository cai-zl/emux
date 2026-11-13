package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/cai-zl/emux/internal/catalog"
)

// groupedHelp renders the root long help: emux's own commands first, then
// the embedded tmux catalog grouped by category.
func groupedHelp() string {
	var b strings.Builder
	b.WriteString(`emux wraps tmux: friendlier common commands, better help, everything else passes straight through to tmux.

Common:
  emux new [name]         create a session          emux ls            list sessions
  emux attach [name]      attach to a session       emux kill [name]   kill (asks to confirm)
  emux rename <old> <new> rename a session          emux detach        detach

tmux commands (` + "`emux help <command>`" + ` for details):
`)
	groups := catalog.ByCategory(catalog.MustLoad())
	tw := tabwriter.NewWriter(&b, 2, 4, 2, ' ', 0)
	for _, cat := range catalog.CategoryOrder {
		fmt.Fprintf(tw, "%s:\n", cat)
		for _, e := range groups[cat] {
			fmt.Fprintf(tw, "  %s\t%s\n", e.Name, e.Description)
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()
	b.WriteString("Anything else is passed to tmux verbatim.\n")
	return b.String()
}

// newHelpCmd replaces cobra's default help command: emux subcommands keep
// cobra help, tmux commands resolve from the catalog, unknown topics pass
// through to tmux's own help.
func newHelpCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "help [command]",
		Short: "show help for a emux or tmux command",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return c.Root().Help()
			}
			for _, sub := range c.Root().Commands() {
				if sub.Name() == args[0] {
					return sub.Help()
				}
			}
			if e, ok := catalog.Lookup(args[0]); ok {
				fmt.Fprintf(c.OutOrStdout(), "%s\n\n%s\n", e.Usage, e.Description)
				return nil
			}
			return a.tmux.Run(append([]string{"help"}, args...)...)
		},
	}
}
