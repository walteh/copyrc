package operation

import (
	"context"

	"github.com/rs/zerolog"
	"gitlab.com/tozd/go/errors"
)

// Clean implements Operator.Clean
func (o *operator) Clean(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Msg("cleaning state")

	// First check if state is consistent
	consistent, err := o.state.IsConsistent(ctx)
	if err != nil {
		return errors.Errorf("checking state consistency: %w", err)
	}
	if !consistent {
		logger.Warn().Msg("state is inconsistent, proceeding with clean")
	}

	// Remove all files tracked by state
	if err := o.state.CleanupOrphanedFiles(ctx); err != nil {
		return errors.Errorf("cleaning up files: %w", err)
	}

	// Reset state to empty
	if err := o.state.Reset(ctx); err != nil {
		return errors.Errorf("resetting state: %w", err)
	}

	logger.Debug().Msg("state cleaned successfully")
	return nil
}

// TODO(dr.methodical): 🧪 Add tests for clean operation
// TODO(dr.methodical): 🧪 Add tests for error cases
// TODO(dr.methodical): 📝 Add example of clean usage
