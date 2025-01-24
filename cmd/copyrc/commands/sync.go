package commands

import (
	"github.com/spf13/cobra"
	"github.com/walteh/copyrc/cmd/copyrc/opts"
	"github.com/walteh/copyrc/pkg/operation"
	"github.com/walteh/copyrc/pkg/remote/github"
	"gitlab.com/tozd/go/errors"
)

// NewSyncCmd creates a new sync command
func NewSyncCmd(opts *opts.RootOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync files from remote repositories",
		Long: `Sync updates the local files with the latest content from remote repositories.
It will:
1. Load the current state
2. Check each repository for changes
3. Download and apply any updates
4. Save the new state`,
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

			// Run sync
			if err := op.Sync(ctx); err != nil {
				return errors.Errorf("syncing files: %w", err)
			}

			return nil
		},
	}

	return cmd
}

// TODO(dr.methodical): 🧪 Add tests for sync command
// TODO(dr.methodical): 🧪 Add tests for provider creation
// TODO(dr.methodical): 📝 Add examples of sync usage
