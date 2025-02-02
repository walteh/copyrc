package main

import "context"

type ProviderFile struct {
	Path     string `json:"path"`
	Dir      bool   `json:"dir"`
	File     bool   `json:"file"`
	Children []ProviderFile
}

type ProviderArgs struct {
	Repo    string
	Ref     string
	Path    string
	RefType string // commit, branch, empty (plain ref)
}

// 🌐 RepoProvider interface for different Git providers
type RepoProvider interface {
	// ListFiles returns a list of files in the given path
	ListFiles(ctx context.Context, args ProviderArgs) ([]ProviderFile, error)
	// GetCommitHash returns the commit hash for the current ref
	GetCommitHash(ctx context.Context, args ProviderArgs) (string, error)
	// GetPermalink returns a permanent link to the file
	GetPermalink(ctx context.Context, args ProviderArgs, commitHash string, file string) (string, error)
	// GetSourceInfo returns a string describing the source (e.g. "github.com/org/repo@hash")
	GetSourceInfo(ctx context.Context, args ProviderArgs, commitHash string) (string, error)
	// GetArchiveUrl returns the URL to download the repository archive
	GetArchiveUrl(ctx context.Context, args ProviderArgs) (string, error)
}
