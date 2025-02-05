archive {
	source {
		repo = "github.com/microsoft/vscode"
		ref  = "tags/1.96.4"
	}
	destination {
		path = "./examples/three/git-repo-tarballs"
	}
	options {
		go_embed = true
	}
}