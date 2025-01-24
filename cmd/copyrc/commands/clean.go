package commands

import (
	"github.com/spf13/cobra"
	"github.com/walteh/copyrc/cmd/copyrc/opts"
	"github.com/walteh/copyrc/pkg/operation"
	"github.com/walteh/copyrc/pkg/remote/github"
	"gitlab.com/tozd/go/errors"
)

// NewCleanCmd creates a new clean command
func NewCleanCmd(opts *opts.RootOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean up local state and files",
		Long: `Clean removes all local state and copied files.
It will:
1. Remove orphaned files
2. Reset the state file
3. Clean up any temporary files`,
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

			// Run clean
			if err := op.Clean(ctx); err != nil {
				return errors.Errorf("cleaning state: %w", err)
			}

			return nil
		},
	}

	return cmd
}

// TODO(dr.methodical): 🧪 Add tests for clean command
// TODO(dr.methodical): 🧪 Add tests for error handling
// TODO(dr.methodical): 📝 Add examples of clean usage
