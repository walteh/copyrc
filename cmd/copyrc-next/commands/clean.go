package commands

import (
	"github.com/spf13/cobra"
	"github.com/walteh/copyrc/cmd/copyrc-next/opts"
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
			return errors.New("not implemented")
		},
	}

	return cmd
}

// TODO(dr.methodical): 🧪 Add tests for clean command
// TODO(dr.methodical): 🧪 Add tests for error handling
// TODO(dr.methodical): 📝 Add examples of clean usage
