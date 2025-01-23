package remote_test

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/walteh/copyrc/pkg/remote"
)

// TestInterfaceMethodSignatures verifies that our interfaces have the correct method signatures
// This is a compile-time check to ensure our interfaces match our requirements
func TestInterfaceMethodSignatures(t *testing.T) {
	// Create type assertions to ensure interfaces have required methods
	var _ interface {
		Name() string
		GetRepository(context.Context, string) (remote.Repository, error)
	} = (remote.Provider)(nil)

	var _ interface {
		Name() string
		GetLatestRelease(context.Context) (remote.Release, error)
		GetReleaseFromRef(context.Context, string) (remote.Release, error)
	} = (remote.Repository)(nil)

	var _ interface {
		Repository() remote.Repository
		Ref() string
		GetTarball(context.Context) (io.ReadCloser, error)
		ListFilesAtPath(context.Context, string) ([]remote.RawTextFile, error)
		GetFileAtPath(context.Context, string) (remote.RawTextFile, error)
		GetLicense(context.Context) (io.ReadCloser, string, error)
		GetLicenseAtPath(context.Context, string) (io.ReadCloser, string, error)
	} = (remote.Release)(nil)

	var _ interface {
		Release() remote.Release
		RawTextPermalink() string
		GetContent(context.Context) (io.ReadCloser, error)
		Path() string
		WebViewPermalink() string
	} = (remote.RawTextFile)(nil)

	assert.True(t, true, "Interface method signatures verified")
}

// TODO(dr.methodical): 🧪 Add mock implementation tests
// TODO(dr.methodical): 🔬 Add integration tests with GitHub implementation
// TODO(dr.methodical): �� Add example tests
