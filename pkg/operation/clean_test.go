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

func TestClean(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mockery.MockStateManager_state, *mockConfig)
		expectedError string
	}{
		{
			name: "successful_clean",
			setupMocks: func(sm *mockery.MockStateManager_state, cfg *mockConfig) {
				sm.EXPECT().IsConsistent(mock.Anything).Return(true, nil)
				sm.EXPECT().CleanupOrphanedFiles(mock.Anything).Return(nil)
				sm.EXPECT().Reset(mock.Anything).Return(nil)
			},
		},
		{
			name: "inconsistent_state_still_cleans",
			setupMocks: func(sm *mockery.MockStateManager_state, cfg *mockConfig) {
				sm.EXPECT().IsConsistent(mock.Anything).Return(false, nil)
				sm.EXPECT().CleanupOrphanedFiles(mock.Anything).Return(nil)
				sm.EXPECT().Reset(mock.Anything).Return(nil)
			},
		},
		{
			name: "consistency_check_error",
			setupMocks: func(sm *mockery.MockStateManager_state, cfg *mockConfig) {
				sm.EXPECT().IsConsistent(mock.Anything).Return(false, assert.AnError)
			},
			expectedError: "checking state consistency",
		},
		{
			name: "cleanup_error",
			setupMocks: func(sm *mockery.MockStateManager_state, cfg *mockConfig) {
				sm.EXPECT().IsConsistent(mock.Anything).Return(true, nil)
				sm.EXPECT().CleanupOrphanedFiles(mock.Anything).Return(assert.AnError)
			},
			expectedError: "cleaning up files",
		},
		{
			name: "reset_error",
			setupMocks: func(sm *mockery.MockStateManager_state, cfg *mockConfig) {
				sm.EXPECT().IsConsistent(mock.Anything).Return(true, nil)
				sm.EXPECT().CleanupOrphanedFiles(mock.Anything).Return(nil)
				sm.EXPECT().Reset(mock.Anything).Return(assert.AnError)
			},
			expectedError: "resetting state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := zerolog.New(zerolog.NewTestWriter(t)).WithContext(context.Background())
			sm := mockery.NewMockStateManager_state(t)
			cfg := &mockConfig{}
			tt.setupMocks(sm, cfg)

			op, err := New(Options{
				Config:       cfg,
				StateManager: sm,
				Provider:     mockery.NewMockProvider_remote(t),
			})
			require.NoError(t, err)

			// Execute
			err = op.Clean(ctx)

			// Verify
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
		})
	}
}

// TODO(dr.methodical): 🧪 Add benchmarks for large state files
// TODO(dr.methodical): 🧪 Add tests for context cancellation
// TODO(dr.methodical): 🧪 Add tests for concurrent access
