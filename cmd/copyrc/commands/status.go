package commands

import (
	"github.com/spf13/cobra"
	"github.com/walteh/copyrc/cmd/copyrc/opts"
	"github.com/walteh/copyrc/pkg/operation"
	"github.com/walteh/copyrc/pkg/remote/github"
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
			ctx := cmd.Context()

			// Create GitHub provider
			provider := github.NewProvider()

			// Create operator
			op, err := operation.New(operation.Options{
				Config:       opts.Config,
				StateManager: opts.StateManager,
				Provider:     provider,
			})
			if err != nil {
				return errors.Errorf("creating operator: %w", err)
			}

			// Run status check
			needsSync, err := op.Status(ctx)
			if err != nil {
				return errors.Errorf("checking status: %w", err)
			}

			// Log result
			if needsSync {
				opts.UserLogger.LogStateChange("Files need to be synced")
			} else {
				opts.UserLogger.LogStateChange("Files are up to date")
			}

			return nil
		},
	}

	return cmd
}

// TODO(dr.methodical): 🧪 Add tests for status command
// TODO(dr.methodical): 🧪 Add tests for error handling
// TODO(dr.methodical): 📝 Add examples of status usage
