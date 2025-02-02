package main

import "context"

type ProviderFile struct {
	Path     string `json:"path"`
	Dir      bool   `json:"dir"`
	File     bool   `json:"file"`
	Children []ProviderFile
}

// 🌐 RepoProvider interface for different Git providers
type RepoProvider interface {
	// ListFiles returns a list of files in the given path
	ListFiles(ctx context.Context, args Source, recursive bool) ([]ProviderFile, error)
	// GetCommitHash returns the commit hash for the current ref
	GetCommitHash(ctx context.Context, args Source) (string, error)
	// GetPermalink returns a permanent link to the file
	GetPermalink(ctx context.Context, args Source, commitHash string, file string) (string, error)
	// GetSourceInfo returns a string describing the source (e.g. "github.com/org/repo@hash")
	GetSourceInfo(ctx context.Context, args Source, commitHash string) (string, error)
	// GetArchiveUrl returns the URL to download the repository archive
	GetArchiveUrl(ctx context.Context, args Source) (string, error)
}
