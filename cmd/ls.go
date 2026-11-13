package cmd

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/cai-zl/emux/internal/tmux"
)

func newLsCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "list sessions",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			ss, err := a.tmux.ListSessions()
			if err != nil {
				return err
			}
			if len(ss) == 0 {
				fmt.Fprintln(c.OutOrStdout(), "no active sessions")
				return nil
			}
			printSessions(c.OutOrStdout(), ss)
			return nil
		},
	}
}

// printSessions renders the shared NAME/WINDOWS/AGE/ATTACHED table used by
// ls and by the list-instead-of-guessing paths of attach/kill/detach.
func printSessions(w io.Writer, ss []tmux.Session) {
	now := time.Now()
	tw := tabwriter.NewWriter(w, 2, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tWINDOWS\tAGE\tATTACHED")
	for _, s := range ss {
		att := "no"
		if s.Attached {
			att = "yes"
		}
		fmt.Fprintf(tw, "%s\t%d\t%s\t%s\n", s.Name, s.Windows, ageOf(s.Created, now), att)
	}
	tw.Flush()
}

func ageOf(created int64, now time.Time) string {
	d := now.Sub(time.Unix(created, 0))
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
