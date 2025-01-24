package operation

import (
	"context"

	"github.com/rs/zerolog"
	"gitlab.com/tozd/go/errors"
)

// Status checks if files need to be fetched from remote
// Returns true if files need to be fetched, false otherwise
func (o *operator) Status(ctx context.Context) (bool, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Msg("checking status")

	// First check if state is consistent
	consistent, err := o.state.IsConsistent(ctx)
	if err != nil {
		return false, errors.Errorf("checking state consistency: %w", err)
	}
	if !consistent {
		logger.Debug().Msg("state is inconsistent")
		return true, nil
	}

	// Then validate local state
	if err := o.state.ValidateLocalState(ctx); err != nil {
		logger.Debug().Err(err).Msg("local state validation failed")
		return true, nil
	}

	// Finally check if config has changed by comparing hashes
	stateHash := o.state.ConfigHash()
	configHash := o.config.Hash()
	if stateHash != configHash {
		logger.Debug().
			Str("state_hash", stateHash).
			Str("config_hash", configHash).
			Msg("config has changed")
		return true, nil
	}

	logger.Debug().Msg("no changes needed")
	return false, nil
}

// TODO(dr.methodical): 🧪 Add tests for different status scenarios
// TODO(dr.methodical): 🧪 Add tests for error cases
// TODO(dr.methodical): 📝 Add example of status usage
