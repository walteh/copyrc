# 📦 copyrc configuration file

repositories {
	provider = "github"
	name     = "google/addlicense"
	ref      = "master"
}

copy {
	repository {
		provider = "github"
		name     = "google/addlicense"
		ref      = "master"
	}
	paths {
		remote = "."
		local  = "pkg/addlicense"
	}
	options {
		text_replacements = [
			{
				from_text        = "package main"
				to_text          = "package addlicense"
				file_filter_glob = "*.go"
			}
		]
		ignore_files_globs = [
			"*.md",
			"go.mod",
			"go.sum",
			"Dockerfile",
			"*.yaml",
			".gitignore",
		]
	}
}

