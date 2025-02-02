copy {
	source {
		repo     = "github.com/omissis/go-jsonschema"
		ref      = "main"
		ref_type = "branch"
		path     = "pkg"
	}
	destination {
		path = "pkg/reformat"
	}
	options {
		recursive = true
		file_patterns = [
			"pkg/generator/**/*.go",
			"pkg/internal/**/*.go",
		]
		replacements = [
			{
				old = "package generator"
				new = "package reformat"
			},
			{
				old = "\"github.com/atombender/go-jsonschema/internal/x/text\""
				new = "\"github.com/walteh/schema2go/pkg/reformat/internal/x/text\""
			}
		]
	}
}

copy {
	source {
		repo     = "github.com/omissis/go-jsonschema"
		ref      = "main"
		ref_type = "branch"
		path     = "tests/data"
	}
	destination {
		path = "pkg/reformat/testdata"
	}
	options {
		recursive = true
		file_patterns = [
			"**/*.go",
			"**/*.json",
		]
		replacements = [
			{
				old = "package test"
				new = "package testdata"
			}
		]
	}
}