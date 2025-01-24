package operation

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/walteh/copyrc/gen/mockery"
)

// mockConfig implements a mock CopyrcConfig for testing
func TestStatus(t *testing.T) {
	tests := []struct {
		name              string
		setupMocks        func(*mockery.MockStateManager_state, *mockery.MockConfig_config)
		expectedNeedsSync bool
		expectedError     string
	}{
		{
			name: "no_changes_needed",
			setupMocks: func(sm *mockery.MockStateManager_state, cfg *mockery.MockConfig_config) {
				sm.EXPECT().IsConsistent(mock.Anything).Return(true, nil)
				sm.EXPECT().ValidateLocalState(mock.Anything).Return(nil)
				sm.EXPECT().ConfigHash().Return("abc123")
				cfg.EXPECT().Hash().Return("abc123")
			},
			expectedNeedsSync: false,
		},
		{
			name: "inconsistent_state",
			setupMocks: func(sm *mockery.MockStateManager_state, cfg *mockery.MockConfig_config) {
				sm.EXPECT().IsConsistent(mock.Anything).Return(false, nil)
			},
			expectedNeedsSync: true,
		},
		{
			name: "invalid_local_state",
			setupMocks: func(sm *mockery.MockStateManager_state, cfg *mockery.MockConfig_config) {
				sm.EXPECT().IsConsistent(mock.Anything).Return(true, nil)
				sm.EXPECT().ValidateLocalState(mock.Anything).Return(assert.AnError)
			},
			expectedNeedsSync: true,
		},
		{
			name: "config_changed",
			setupMocks: func(sm *mockery.MockStateManager_state, cfg *mockery.MockConfig_config) {
				sm.EXPECT().IsConsistent(mock.Anything).Return(true, nil)
				sm.EXPECT().ValidateLocalState(mock.Anything).Return(nil)
				sm.EXPECT().ConfigHash().Return("abc123")
				cfg.EXPECT().Hash().Return("def456")
			},
			expectedNeedsSync: true,
		},
		{
			name: "consistency_check_error",
			setupMocks: func(sm *mockery.MockStateManager_state, cfg *mockery.MockConfig_config) {
				sm.EXPECT().IsConsistent(mock.Anything).Return(false, assert.AnError)
			},
			expectedError: "checking state consistency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := zerolog.New(zerolog.NewTestWriter(t)).WithContext(context.Background())
			sm := mockery.NewMockStateManager_state(t)
			cfg := mockery.NewMockConfig_config(t)
			tt.setupMocks(sm, cfg)

			op, err := New(Options{
				Config:       cfg,
				StateManager: sm,
				Provider:     mockery.NewMockProvider_remote(t),
			})
			require.NoError(t, err)

			// Execute
			needsSync, err := op.Status(ctx)

			// Verify
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedNeedsSync, needsSync)
		})
	}
}

// TODO(dr.methodical): 🧪 Add benchmarks for large state files
// TODO(dr.methodical): 🧪 Add tests for context cancellation
// TODO(dr.methodical): 🧪 Add tests for concurrent access
