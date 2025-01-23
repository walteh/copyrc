/*
  This file is a guiding specification describing how the
  refactored "copyrc-legacy" codebase should be organized.

  NOTE: This isn't code to be parsed at runtime; it's a "living blueprint"
  to help define scope and structure for the refactor.
*/

packages {
	remote {
		nested_implementation_packages {
			github {
				implementation_notes = <<EOT
					- All GitHub API calls should use context.Context for timeouts/cancellation
					- Use github.com/google/go-github/v60 client library
					- ONLY cache the tarballs for each release, not the api responses
					- use the tarballs to get the files
				EOT
			}
		}
		interfaces {
			Provider {
				# PROVIDE_MORE_CLARITY: Should this interface handle rate limiting and retries?
				description = "Primary interface for interacting with remote repository providers (e.g. GitHub)"
				methods = {
					"Name() -> string"                                = true
					"GetRepository(ctx, name) -> (Repository, error)" = true
				}
			}
			Repository {
				description = "Represents a remote code repository and its releases"
				methods = {
					"Name() -> string"                                = true
					"GetLatestRelease(ctx) -> (Release, error)"       = true
					"GetReleaseFromRef(ctx, ref) -> (Release, error)" = true
				}
			}
			Release {
				description = "Represents a specific version/tag/commit of a repository"
				methods = {
					"Repository() -> Repository"                                  = true
					"Ref() -> string"                                             = true
					"GetTarball(ctx) -> (io.ReadCloser, error)"                   = true
					"ListFilesAtPath(ctx, path) -> ([]RawTextFile, error)"        = true
					"GetFileAtPath(ctx, path) -> (RawTextFile, error)"            = true
					"GetLicense(ctx) -> (io.ReadCloser, spdx, error)"             = true # spdx is just the license type string
					"GetLicenseAtPath(ctx, path) -> (io.ReadCloser, spdx, error)" = true
				}
			}
			RawTextFile {
				description = "Represents a text file from a specific release that can be downloaded"
				methods = {
					"Release() -> Release"                      = true
					"RawTextPermalink() -> string"              = true
					"GetContent(ctx) -> (io.ReadCloser, error)" = true
					"Path() -> string"                          = true
					"WebViewPermalink() -> string"              = true
				}
			}
		}
	}
	config {
		description = "Package for loading and validating user configuration"
		structs {
			CopyrcConfig {
				description = "Top-level configuration structure defining what to copy from where"
				fields {
					Repositories = "[]RepositoryDefinition"
					Copies       = "[]Copy"
				}
				methods = {
					"Hash() -> string" = true # Used to detect config changes
				}
			}
			Copy {
				description = "Defines a single copy operation from a remote source to local destination"
				fields {
					repository = "RepositoryDefinition"
					paths = { # can be directories or files
						remote = "string" # Path in remote repository
						local  = "string" # Path on local filesystem
					}
					options {
						# PROVIDE_MORE_CLARITY: What is the format of Replace? Need struct definition
						text_replacements           = "[]Replace"
						save_archive_to_path        = "string"
						create_go_embed_for_archive = "bool"
					}
				}
			}
			RepositoryDefinition {
				description = "Identifies a specific remote repository and version"
				fields {
					provider = "string" # e.g. "github"
					name     = "string" # e.g. "owner/repo"
					ref      = "string" # tag, branch, or commit
				}
			}
		}
	}
	state {
		description = <<EOT
			Package for managing local state of copied files.
			- Uses .copyrc.lock JSON files to track file status
			- One lock file per .copyrc definition
			- Handles patching and file modifications
			- Provides idempotent operations
		EOT

		implementation_notes = <<EOT
			File State Management:
			---------------------
			1. Remote vs Local:
			   - Remote package manages remote files
			   - State package manages local files via .copyrc.lock JSON files
			   - Each directory with copied files has its own lock file

			2. Idempotent Behavior:
			   - When another package notices file state changes, state package applies them
			   - Example: operation package sends remote file content without comparison
			   - State package determines if local file needs updating

			3. File Tracking & Cleanup:
			   - Must track all files it was notified about
			   - Handles file rename/removal scenarios
			   - Example: if "foo.txt" becomes "foo.md", state package must:
			     * Detect "foo.txt" is no longer needed
			     * Remove "foo.txt"
			     * Track "foo.md" as new file
			   - Implementation suggestion: Use boolean field for tracking updates
			     * Set to true when file is updated
			     * Check for false values to identify files for deletion

			4. Interface Design:
			   - Main State struct should not be used directly by other packages
			   - Instead, expose focused interfaces for specific operations
			   - This allows better control over state modifications

			5. Logging Requirements:
			   - Implement robust logging system
			   - Log all actions taken on files
			   - Provide clear user feedback about:
			     * File updates
			     * Deletions
			     * State changes
			     * Errors
		EOT

		structs {
			State {
				description = "Top-level state tracking structure that manages file state and cleanup"
				fields {
					last_updated = "string"
					repositories = {
						provider   = "string"
						name       = "string" # also the unique identifier for the repository
						latest_ref = "string"
						target_release = {
							last_updated = "string"
							ref          = "string" # also the unique identifier for the release
							ref_hash     = "string"
							archive = {
								hash         = "string"
								content_type = "string"
								download_url = "string"
								local_path   = "string" # might not be saved, only if save_archive_to_path
							}
							web_permalink = "string"
							license = {
								spdx             = "string"
								remote_permalink = "string"
								local_path       = "string" # might not be saved, only if save_archive_to_path -
							}

						}
					}
					remote_text_files = "[]RemoteTextFile"
					generated_files   = "[]GeneratedFile"
					config            = "Copy"
				}
				methods = {
					description = "Methods for updating state and tracking files"
					"PutRemoteTextFile(ctx, provider.RawTextFile, local_path) -> (RemoteTextFile, error)" = true
					"PutGeneratedFile(ctx, GeneratedFile) -> (GeneratedFile, error)"                      = true
					"PutArchiveFile(ctx, provider.Release, local_path) -> (ArchiveFile, error)"           = true
				}
			}
			GeneratedFile { # right now the only generated file is the go embed for the tarball
				fields {
					local_path     = "string"
					last_updated   = "string"
					reference_file = "string"
				}
			}
			RemoteTextFile {
				fields {
					# PROVIDE_MORE_CLARITY: What specific metadata fields are expected/required?
					metadata = "map[string]string"
					repository_name = "string"
					release_ref     = "string"
					# all saved files files should have the same name as the remote file, but with the .copy.xyz or .patch.xyz suffix. ex "foo.txt" -> "foo.copy.txt" or "foo.patch.txt"
					local_path          = "string" # will be either the local path to the raw file ".copy." or to a patch ".patch."
					last_updated        = "string"
					is_patched          = "bool"
					remote_content_hash = "string"
					patch {
						# PROVIDE_MORE_CLARITY: What format should the patch_diffs follow? 
						patch_diff = "string" # gopatch diff format
						gzipped_remote_content = "string"
						patch_path             = "string"
					}
					permalink                          = "string"
					auto_text_replacment_modifications = "[]AutoTextReplacementModification"
					license = {
						spdx             = "string"
						local_path       = "string"
						remote_permalink = "string"
					}
				}
				methods = {
					# when the file is loaded into memory, it can pull in the contents of these files
					# this content is saved to memory along with the writing of the state file
					"RawRemoteContent() -> (io.ReadCloser, error)"                      = true # this may not need to be saved
					"RawPatchContent() -> (io.ReadCloser, error)"                       = true # this may not need to be saved
					"ApplyModificationToRawRemoteContent(ctx, Modification) -> (error)" = true

				}
			}
			AutoTextReplacementModification {
				fields {
					from_text      = "string"
					to_text        = "string"
					file_filter_glob = "string"
				}
			}
		}

		methods = {
			"LoadState(ctx, path) -> (State, error)" = true
			"WriteState(ctx, path) -> (error)"       = true

		}
	}
	operation {
		# PROVIDE_MORE_CLARITY: Need specific implementation details about orchestration
		description = <<EOT
			Package for coordinating high-level operations between config, remote, and state packages.
			Responsible for:
			- Reading config
			- Fetching remote content
			- Updating local state
			- Applying patches and modifications
		EOT

		implementation_notes = <<EOT
			Ensure all operations are idempotent and handle errors gracefully
			Implement logging to provide user feedback on operations
			Use interfaces to decouple operations from specific implementations

			the operations that it needs to process are 
				"sync" which just updates the state file with the latest remote file content
				"clean" which clears the state
				"status" which is a local operation indicating if the files need to be fetched from remote (i.e. the config has changed)
		EOT
	}

}

standards {
	error_handling {
		library = "gitlab.com/tozd/go/errors"
		style   = <<EOT
         Always wrap low-level errors with context about what we were trying to do,
         e.g. errors.Errorf("fetching release data: %w", err)
       EOT
	}
	debug_logging {
		library = "github.com/rs/zerolog"
		style   = <<EOT
			All methods should receive a context and refer to the logger as zerolog.Ctx(ctx)
			don't ever store the logger in a variable, always use the zerolog.Ctx(ctx) method
			this is traditional application logging, nothing fancy
		EOT
	}
	user_logging {
		library = "github.com/pterm/pterm"
		style   = <<EOT
			we need to define our own custom logging for users that shows them nice colored progress bars and shoes them which files are being downloaded and which files are being copied, which are patched, etc
			this is an entirely different logging system than the debug logging system
		EOT
	}
	testing {
		library = "github.com/stretchr/testify"
		style   = <<EOT
			Table-driven tests for each method, with mocks for I/O or remote calls.
			unless ABSOLUTELY NECESSARY, all tests should be in a _test package
			For interfaces that are required by a package, create a mock in the .mockery.yaml file and import the mock to the test
		EOT
	}
	patching_files {
		library = "github.com/uber-go/gopatch"
		info <<EOT
			unfortuantly, this package is all internal - so for now we should just use it as a command line tool to generate the patch files
			read https://raw.githubusercontent.com/uber-go/gopatch/refs/heads/main/docs/PatchesInDepth.md for more information

			eventually we will use copyrc to download the content into our repo, but for now lets just use the gopatch command line tool to generate the patch files
			use it by running go run github.com/uber-go/gopatch -f patch.diff -o patch.patch
		EOT
	}
}


notes_thoughts_and_clarifications {
	license_files {
		info <<EOT
			we are just saving the licenses along with any coppied code. each root directly of copied code needs a single license
			the license file being used should be the one that is closest in the remote directory root to the copied code 

		EOT
	}

	downloading_files {
		info <<EOT
			initailly i had expected we would entirly use the github api, but after looking at this doc it seems we could probably take advantage of those 
			archive tars. like download them for each release and then just use the files from the tarball. this would allow us to not have to deal with the github api at all other than to get the download url for the tarball
		EOT
	}	

	integration_tests {
		info <<EOT
			we should have a test that downloads the license file, then the copied code, then the patch file and then checks the files are in the state file
			also, we should make sure any generated link we create has a test to make sure it connects to github and is correct
		EOT
	}

	logging {
		info <<EOT
			we need two different logging systems, one for debugging and one for displaying updates to the user. 
			we should use the zerolog library for the debugging / testing logging system
			we need to define our own custom logging 
		EOT
	}
}


questions {
	a = {
		question = <<EOT
			what is the format of Replace?
		EOT
		answer = <<EOT
			Replace is a struct with the following fields: from_text, to_text, line_number, start_position, end_position
		EOT
	}

	b = {
		question = <<EOT
			What specific caching strategy should be used for GitHub API responses?
			
			Context:
			- The implementation notes mention caching "when practical"
			- Need to specify:
			  * Cache storage (memory/disk)
			  * Cache duration
			  * Cache invalidation rules
			  * Cache size limits
		EOT
		answer = <<EOT
			tar archives that are downloaded can be cached in the file system for a specific tag. we need an invalidation operation to remove the cache 
			but we should nto cache the github api responses, the "status" operation should help us prevent extra calls to the github api
		EOT
	}

	c = {
		question = <<EOT
			Should the Provider interface include rate limiting and retry logic?
			
			Context:
			- GitHub has API rate limits
			- Need to specify:
			  * Should retries be automatic?
			  * How many retries?
			  * Should backoff be exponential?
			  * Should rate limits be shared across Provider instances?
		EOT
		answer = <<EOT
			throw an error if the github api rate limit is exceeded, tell the user to add a github token if they want to increase the rate limit
		EOT
	}

	d = {
		question = <<EOT
			What format should license handling follow?
			
			Context:
			- Methods return (io.ReadCloser, spdx, error)
			- Need to specify:
			  * SPDX format requirements
			  * License content validation
			  * How to handle multiple license files
			  * Priority rules for license selection
		EOT
		answer = <<EOT
			 no need to validate the content of the license file, its just to keep track of it as needed
		EOT
	}

	e = {
		question = <<EOT
			What metadata fields are expected/required for RemoteTextFile?
			
			Context:
			- Currently defined as map[string]string
			- Need to specify:
			  * Required fields
			  * Optional fields
			  * Field validation rules
			  * Field format requirements
		EOT
		answer = <<EOT
			not clear right now, ideally it would be stuff specific to the provider, so github specific swtuff for now. its there just a vessle for unexpected data
		EOT
	}

	f = {
		question = <<EOT
			What format should patch_diffs follow?
			
			Context:
			- Part of RemoteTextFile struct
			- Need to specify:
			  * Patch format standard
			  * Conflict resolution rules
			  * Atomic application requirements
			  * Validation requirements
		EOT
		answer = <<EOT
			use gopatch to generate the patch files, for now that is a command line tool, in the future we will use copyrc to download the content into our repo
		EOT
	}



}