copy {
	source {
		repo = "github.com/golang/tools"
		ref  = "master"
		path = "gopls/internal/protocol/generate"
	}
	destination {
		path = "./examples/dirs/gopls-protocol-generator"
	}
	options {
		replacements = [
			{
				old = "tsprotocol.go",
				new = "tsprotocol.gen.go"
			},
			{
				old = "tsserver.go",
				new = "tsserver.gen.go"
			},
			{
				old = "tsclient.go",
				new = "tsclient.gen.go"
			},
			{
				old = "tsjson.go",
				new = "tsjson.gen.go"
			}
		]
		ignore_files = [
			"*.txt",
		]
		file_patterns = [
			"*.go",      # Only copy Go files
			"*.js",      # And JavaScript files
			"docs/*.md", # And markdown files in docs directory
		]
	}
}

archive {
	source {
		repo = "github.com/neovim/nvim-lspconfig"
		ref  = "tags/v1.3.0"
	}
	destination {
		path = "./examples/tarballs"
	}
	options {
		go_embed = true
	}
}

archive {
	source {
		repo = "github.com/microsoft/vscode-languageserver-node"
		ref  = "tags/release/jsonrpc/9.0.0-next.6"
	}
	destination {
		path = "./examples/tarballs"
	}
	options {
		go_embed = true
	}
}

