package opts

import (
	"github.com/walteh/copyrc/pkg/config"
	"github.com/walteh/copyrc/pkg/state"
)

// RootOpts contains shared options used by all commands
type RootOpts struct {
	Config       *config.CopyrcConfig
	StateManager state.StateManager
	UserLogger   *state.UserLogger
}

// TODO(dr.methodical): 🧪 Add tests for option validation
// TODO(dr.methodical): 📝 Add examples of option usage
