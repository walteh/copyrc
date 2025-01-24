package commands

import (
	"github.com/spf13/cobra"
	"github.com/walteh/copyrc/cmd/copyrc-next/opts"
	"gitlab.com/tozd/go/errors"
)

// NewStatusCmd creates a new status command
func NewStatusCmd(opts *opts.RootOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check if files need to be synced",
		Long: `Status checks if local files are up to date with remote repositories.
It will:
1. Load the current state
2. Check state consistency
3. Compare configuration hashes
4. Report if files need to be synced`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("not implemented")
		},
	}

	return cmd
}

// TODO(dr.methodical): 🧪 Add tests for status command
// TODO(dr.methodical): 🧪 Add tests for error handling
// TODO(dr.methodical): 📝 Add examples of status usage
