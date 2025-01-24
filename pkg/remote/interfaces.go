package remote

import (
	"context"
	"io"
	"strings"

	"github.com/walteh/copyrc/pkg/config"
	"gitlab.com/tozd/go/errors"
)

var registry = map[string]Provider{}

func RegisterProvider(name string, provider Provider) {
	registry[name] = provider
}

func GetProviderFromConfig(ctx context.Context, cn config.RepositoryDefinition) (Provider, error) {
	provider, ok := registry[cn.Provider]
	if !ok {
		options := []string{}
		for k := range registry {
			options = append(options, k)
		}
		return nil, errors.Errorf("provider %s not found, options: %s", cn.Provider, strings.Join(options, ", "))
	}
	return provider, nil
}

// Provider is the primary interface for interacting with remote repository providers (e.g. GitHub)
type Provider interface {
	// Name returns the name of the provider (e.g. "github")
	Name() string
	// GetRepository returns a Repository interface for the given repository name
	GetRepository(ctx context.Context, name string) (Repository, error)
}

// Repository represents a remote code repository and its releases
type Repository interface {
	// Name returns the name of the repository (e.g. "owner/repo")
	Name() string
	// GetLatestRelease returns the latest release for this repository
	GetLatestRelease(ctx context.Context) (Release, error)
	// GetReleaseFromRef returns a specific release for this repository
	GetReleaseFromRef(ctx context.Context, ref string) (Release, error)
}

// Release represents a specific version/tag/commit of a repository
type Release interface {
	// Repository returns the parent repository
	Repository() Repository
	// Ref returns the reference (tag/branch/commit) for this release
	Ref() string
	// GetTarball returns a reader for the tarball of this release
	GetTarball(ctx context.Context) (io.ReadCloser, error)
	// ListFilesAtPath lists all files at a given path in this release
	ListFilesAtPath(ctx context.Context, path string) ([]RawTextFile, error)
	// GetFileAtPath returns a specific file from this release
	GetFileAtPath(ctx context.Context, path string) (RawTextFile, error)
	// GetLicense returns the license file for this release
	GetLicense(ctx context.Context) (License, error)
	// GetLicenseAtPath returns a license file at a specific path
	GetLicenseAtPath(ctx context.Context, path string) (License, error)
	// WebPermalink returns a permanent link to view the repository on the web
	WebPermalink() string
	RefHash() string
}

// RawTextFile represents a text file from a specific release that can be downloaded
type RawTextFile interface {
	// Release returns the parent release
	Release() Release
	// RawTextPermalink returns a permanent link to the raw text content
	RawTextPermalink() string
	// GetContent returns a reader for the file content
	GetContent(ctx context.Context) (io.ReadCloser, error)
	// Path returns the path of the file in the repository
	Path() string
	// WebViewPermalink returns a permanent link to view the file on the web
	WebViewPermalink() string
}

// ModifiableRawTextFile extends RawTextFile with the ability to modify its content
type ModifiableRawTextFile interface {
	RawTextFile
	SetContent(content string)
}

type License struct {
	SPDXID       string
	WebPermalink string
}

// TODO(dr.methodical): 🔬 Add mockery configuration for these interfaces
// TODO(dr.methodical): 🧪 Add tests for interface method signatures
// TODO(dr.methodical): 🎯 Add GitHub implementation
// TODO(dr.methodical):  Add godoc examples
