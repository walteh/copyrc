copy {
	source {
		repo = "github.com/google/addlicense"
		ref  = "master"
		path = "."
	}

	destination {
		path = "pkg/addlicense"
	}

	options {
		replacements = [
			{
				old = "package main"
				new = "package addlicense"
			}
		]
		ignore_files = [
			"README.md",
			"go.mod",
			"go.sum",
			"go.work",
			"go.work.sum",
			"*.yaml",
			"Dockerfile",
		]
	}
}

