// Copied from github.com/daytonaio/daytona/libs/sdk-go/pkg/options v0.190.0 (Apache-2.0); see NOTICE.

package options

import (
	types "github.com/digitalocean/godo/sandbox/compat/daytona/types"
	time "time"
)

type CreateSandbox struct {
	Timeout      *time.Duration // Maximum time to wait for sandbox creation
	WaitForStart bool           // Whether to wait for the sandbox to reach started state
	LogChannel   chan string    // Channel for receiving build logs during image builds
}

// WithTimeout sets the maximum duration to wait for sandbox creation to complete.
//
// If the timeout is exceeded before the sandbox is ready, Create returns an error.
// The default timeout is 60 seconds.
//
// Example:
//
//	sandbox, err := client.Create(ctx, params,
//	    options.WithTimeout(5*time.Minute),
//	)
func WithTimeout(timeout time.Duration) func(*CreateSandbox) {
	return func(opts *CreateSandbox) {
		opts.Timeout = &timeout
	}
}

// WithWaitForStart controls whether [daytona.Client.Create] waits for the sandbox
// to reach the started state before returning.
//
// When true (the default), Create blocks until the sandbox is fully started and ready
// for use. When false, Create returns immediately after the sandbox is created,
// which may be in a pending or building state.
//
// Example:
//
//	// Return immediately without waiting for the sandbox to start
//	sandbox, err := client.Create(ctx, params,
//	    options.WithWaitForStart(false),
//	)
func WithWaitForStart(waitForStart bool) func(*CreateSandbox) {
	return func(opts *CreateSandbox) {
		opts.WaitForStart = waitForStart
	}
}

// WithLogChannel provides a channel for receiving build logs during sandbox creation.
//
// When creating a sandbox from a custom image that requires building, build logs
// are streamed to the provided channel. The channel is closed when streaming completes.
// If no build is required, no logs are sent and the channel remains unused.
//
// Example:
//
//	logChan := make(chan string)
//	go func() {
//	    for log := range logChan {
//	        fmt.Println(log)
//	    }
//	}()
//	sandbox, err := client.Create(ctx, params,
//	    options.WithLogChannel(logChan),
//	)
func WithLogChannel(logChannel chan string) func(*CreateSandbox) {
	return func(opts *CreateSandbox) {
		opts.LogChannel = logChannel
	}
}

type RunCode struct {
	ContextID string            // Interpreter context ID for persistent state
	Env       map[string]string // Environment variables for code execution
	Timeout   *time.Duration    // Execution timeout
}

// WithCustomContext sets the interpreter context ID for code execution.
//
// Using a context allows you to maintain state (variables, imports, etc.)
// across multiple code executions. Create a context with CreateContext first.
//
// Example:
//
//	ctx, _ := sandbox.CodeInterpreter.CreateContext(ctx, nil)
//	channels, err := sandbox.CodeInterpreter.RunCode(ctx, "x = 42",
//	    options.WithCustomContext(ctx["id"].(string)),
//	)
func WithCustomContext(contextID string) func(*RunCode) {
	return func(opts *RunCode) {
		opts.ContextID = contextID
	}
}

// WithEnv sets environment variables for code execution.
//
// These variables are available to the code during execution.
//
// Example:
//
//	channels, err := sandbox.CodeInterpreter.RunCode(ctx, "import os; print(os.environ['API_KEY'])",
//	    options.WithEnv(map[string]string{"API_KEY": "secret"}),
//	)
func WithEnv(env map[string]string) func(*RunCode) {
	return func(opts *RunCode) {
		opts.Env = env
	}
}

// WithInterpreterTimeout sets the execution timeout for code.
//
// If the code doesn't complete within the timeout, execution is terminated.
//
// Example:
//
//	channels, err := sandbox.CodeInterpreter.RunCode(ctx, "import time; time.sleep(60)",
//	    options.WithInterpreterTimeout(5*time.Second),
//	)
func WithInterpreterTimeout(timeout time.Duration) func(*RunCode) {
	return func(opts *RunCode) {
		opts.Timeout = &timeout
	}
}

type CreateFolder struct {
	Mode *string // Unix file permissions (e.g., "0755")
}

// WithMode sets the Unix file permissions for the created folder.
//
// The mode should be specified as an octal string (e.g., "0755", "0700").
// If not specified, defaults to "0755".
//
// Example:
//
//	err := sandbox.FileSystem.CreateFolder(ctx, "/home/user/mydir",
//	    options.WithMode("0700"),
//	)
func WithMode(mode string) func(*CreateFolder) {
	return func(opts *CreateFolder) {
		opts.Mode = &mode
	}
}

type SetFilePermissions struct {
	Mode  *string // Unix file permissions (e.g., "0644")
	Owner *string // File owner username
	Group *string // File group name
}

// WithPermissionMode sets the Unix file permissions.
//
// The mode should be specified as an octal string (e.g., "0644", "0755").
//
// Example:
//
//	err := sandbox.FileSystem.SetFilePermissions(ctx, "/home/user/file.txt",
//	    options.WithPermissionMode("0644"),
//	)
func WithPermissionMode(mode string) func(*SetFilePermissions) {
	return func(opts *SetFilePermissions) {
		opts.Mode = &mode
	}
}

// WithOwner sets the file owner.
//
// The owner should be a valid username on the sandbox system.
//
// Example:
//
//	err := sandbox.FileSystem.SetFilePermissions(ctx, "/home/user/file.txt",
//	    options.WithOwner("root"),
//	)
func WithOwner(owner string) func(*SetFilePermissions) {
	return func(opts *SetFilePermissions) {
		opts.Owner = &owner
	}
}

// WithGroup sets the file group.
//
// The group should be a valid group name on the sandbox system.
//
// Example:
//
//	err := sandbox.FileSystem.SetFilePermissions(ctx, "/home/user/file.txt",
//	    options.WithGroup("users"),
//	)
func WithGroup(group string) func(*SetFilePermissions) {
	return func(opts *SetFilePermissions) {
		opts.Group = &group
	}
}

// Apply creates a new instance of type T and applies all provided option functions.
//
// This generic function enables a consistent pattern for applying functional options
// across different option types. It allocates a zero-value instance of T, then
// applies each option function in order.
//
// Example:
//
//	opts := options.Apply(
//	    options.WithBranch("main"),
//	    options.WithUsername("user"),
//	)
func Apply[T any](opts ...func(*T)) *T {
	result := new(T)
	for _, opt := range opts {
		opt(result)
	}
	return result
}

type GitClone struct {
	Branch          *string // Branch to clone (defaults to repository's default branch)
	CommitId        *string // Specific commit SHA to checkout after cloning
	Username        *string // Username for HTTPS authentication
	Password        *string // Password or token for HTTPS authentication
	InsecureSkipTLS *bool   // Skip TLS certificate verification (insecure). Use only for trusted internal Git servers with self-signed or private-CA certs.
}

// WithBranch sets the branch to clone instead of the repository's default branch.
//
// Example:
//
//	err := sandbox.Git.Clone(ctx, url, path, options.WithBranch("develop"))
func WithBranch(branch string) func(*GitClone) {
	return func(opts *GitClone) {
		opts.Branch = &branch
	}
}

// WithCommitId sets a specific commit SHA to checkout after cloning.
//
// The repository is first cloned, then the specified commit is checked out,
// resulting in a detached HEAD state.
//
// Example:
//
//	err := sandbox.Git.Clone(ctx, url, path, options.WithCommitId("abc123def"))
func WithCommitId(commitID string) func(*GitClone) {
	return func(opts *GitClone) {
		opts.CommitId = &commitID
	}
}

// WithUsername sets the username for HTTPS authentication when cloning.
//
// For GitHub, GitLab, and similar services, the username is typically your
// account username or a placeholder like "git" when using tokens.
//
// Example:
//
//	err := sandbox.Git.Clone(ctx, url, path,
//	    options.WithUsername("username"),
//	    options.WithPassword("github_token"),
//	)
func WithUsername(username string) func(*GitClone) {
	return func(opts *GitClone) {
		opts.Username = &username
	}
}

// WithPassword sets the password or access token for HTTPS authentication when cloning.
//
// For GitHub, use a Personal Access Token (PAT). For GitLab, use a Project
// Access Token or Personal Access Token. For Bitbucket, use an App Password.
//
// Example:
//
//	err := sandbox.Git.Clone(ctx, url, path,
//	    options.WithUsername("username"),
//	    options.WithPassword("ghp_xxxxxxxxxxxx"),
//	)
func WithPassword(password string) func(*GitClone) {
	return func(opts *GitClone) {
		opts.Password = &password
	}
}

// WithInsecureSkipTLS opts into skipping TLS certificate verification for the
// clone. Use ONLY when cloning from a trusted internal Git server with a
// self-signed or private-CA certificate. Credentials, if supplied, will be
// transmitted over an unverified TLS connection and are exposed to any MITM
// on the route.
//
// Example:
//
//	err := sandbox.Git.Clone(ctx, url, path,
//	    options.WithInsecureSkipTLS(true),
//	)
func WithInsecureSkipTLS(insecureSkipTLS bool) func(*GitClone) {
	return func(opts *GitClone) {
		opts.InsecureSkipTLS = &insecureSkipTLS
	}
}

type GitCommit struct {
	AllowEmpty *bool // Allow creating commits with no staged changes
}

// WithAllowEmpty allows creating a commit even when there are no staged changes.
//
// This is useful for triggering CI/CD pipelines or marking points in history
// without actual code changes.
//
// Example:
//
//	resp, err := sandbox.Git.Commit(ctx, path, "Trigger rebuild", author, email,
//	    options.WithAllowEmpty(true),
//	)
func WithAllowEmpty(allowEmpty bool) func(*GitCommit) {
	return func(opts *GitCommit) {
		opts.AllowEmpty = &allowEmpty
	}
}

type GitDeleteBranch struct {
	Force *bool // Force delete even if branch is not fully merged
}

// WithForce enables force deletion of a branch even if it's not fully merged.
//
// Use with caution as this can result in lost commits if the branch contains
// work that hasn't been merged elsewhere.
//
// Example:
//
//	err := sandbox.Git.DeleteBranch(ctx, path, "feature/abandoned",
//	    options.WithForce(true),
//	)
func WithForce(force bool) func(*GitDeleteBranch) {
	return func(opts *GitDeleteBranch) {
		opts.Force = &force
	}
}

type GitPush struct {
	Username *string // Username for HTTPS authentication
	Password *string // Password or token for HTTPS authentication
}

// WithPushUsername sets the username for HTTPS authentication when pushing.
//
// Example:
//
//	err := sandbox.Git.Push(ctx, path,
//	    options.WithPushUsername("username"),
//	    options.WithPushPassword("github_token"),
//	)
func WithPushUsername(username string) func(*GitPush) {
	return func(opts *GitPush) {
		opts.Username = &username
	}
}

// WithPushPassword sets the password or access token for HTTPS authentication when pushing.
//
// Example:
//
//	err := sandbox.Git.Push(ctx, path,
//	    options.WithPushUsername("username"),
//	    options.WithPushPassword("ghp_xxxxxxxxxxxx"),
//	)
func WithPushPassword(password string) func(*GitPush) {
	return func(opts *GitPush) {
		opts.Password = &password
	}
}

type GitPull struct {
	Username *string // Username for HTTPS authentication
	Password *string // Password or token for HTTPS authentication
}

// WithPullUsername sets the username for HTTPS authentication when pulling.
//
// Example:
//
//	err := sandbox.Git.Pull(ctx, path,
//	    options.WithPullUsername("username"),
//	    options.WithPullPassword("github_token"),
//	)
func WithPullUsername(username string) func(*GitPull) {
	return func(opts *GitPull) {
		opts.Username = &username
	}
}

// WithPullPassword sets the password or access token for HTTPS authentication when pulling.
//
// Example:
//
//	err := sandbox.Git.Pull(ctx, path,
//	    options.WithPullUsername("username"),
//	    options.WithPullPassword("ghp_xxxxxxxxxxxx"),
//	)
func WithPullPassword(password string) func(*GitPull) {
	return func(opts *GitPull) {
		opts.Password = &password
	}
}

type PipInstall struct {
	FindLinks      []string // URLs to search for packages
	IndexURL       string   // Base URL of the Python Package Index
	ExtraIndexURLs []string // Extra index URLs for package lookup
	Pre            bool     // Allow pre-release and development versions
	ExtraOptions   string   // Additional pip command-line options
}

// WithFindLinks adds find-links URLs for pip install.
//
// Find-links URLs are searched for packages before the package index.
// Useful for installing packages from local directories or custom URLs.
//
// Example:
//
//	image := daytona.Base("python:3.11").PipInstall(
//	    []string{"mypackage"},
//	    options.WithFindLinks("/path/to/wheels", "https://example.com/wheels/"),
//	)
func WithFindLinks(links ...string) func(*PipInstall) {
	return func(opts *PipInstall) {
		opts.FindLinks = append(opts.FindLinks, links...)
	}
}

// WithIndexURL sets the base URL of the Python Package Index.
//
// Replaces the default PyPI (https://pypi.org/simple) with a custom index.
//
// Example:
//
//	image := daytona.Base("python:3.11").PipInstall(
//	    []string{"mypackage"},
//	    options.WithIndexURL("https://my-pypi.example.com/simple/"),
//	)
func WithIndexURL(url string) func(*PipInstall) {
	return func(opts *PipInstall) {
		opts.IndexURL = url
	}
}

// WithExtraIndexURLs adds extra index URLs for pip install.
//
// Extra indexes are checked in addition to the main index URL.
// Useful for installing packages from both PyPI and a private index.
//
// Example:
//
//	image := daytona.Base("python:3.11").PipInstall(
//	    []string{"mypackage"},
//	    options.WithExtraIndexURLs("https://private.example.com/simple/"),
//	)
func WithExtraIndexURLs(urls ...string) func(*PipInstall) {
	return func(opts *PipInstall) {
		opts.ExtraIndexURLs = append(opts.ExtraIndexURLs, urls...)
	}
}

// WithPre enables installation of pre-release and development versions.
//
// Example:
//
//	image := daytona.Base("python:3.11").PipInstall(
//	    []string{"mypackage"},
//	    options.WithPre(),
//	)
func WithPre() func(*PipInstall) {
	return func(opts *PipInstall) {
		opts.Pre = true
	}
}

// WithExtraOptions adds extra command-line options for pip install.
//
// Use this for pip options not covered by other With* functions.
//
// Example:
//
//	image := daytona.Base("python:3.11").PipInstall(
//	    []string{"mypackage"},
//	    options.WithExtraOptions("--no-cache-dir --upgrade"),
//	)
func WithExtraOptions(options string) func(*PipInstall) {
	return func(opts *PipInstall) {
		opts.ExtraOptions = options
	}
}

type ExecuteCommand struct {
	Cwd     *string           // Working directory for command execution
	Env     map[string]string // Environment variables
	Timeout *time.Duration    // Command execution timeout
}

// WithCwd sets the working directory for command execution.
//
// Example:
//
//	result, err := sandbox.Process.ExecuteCommand(ctx, "ls -la",
//	    options.WithCwd("/home/user/project"),
//	)
func WithCwd(cwd string) func(*ExecuteCommand) {
	return func(opts *ExecuteCommand) {
		opts.Cwd = &cwd
	}
}

// WithCommandEnv sets environment variables for the command.
//
// These variables are added to the command's environment in addition to
// the sandbox's default environment.
//
// Example:
//
//	result, err := sandbox.Process.ExecuteCommand(ctx, "echo $MY_VAR",
//	    options.WithCommandEnv(map[string]string{"MY_VAR": "hello"}),
//	)
func WithCommandEnv(env map[string]string) func(*ExecuteCommand) {
	return func(opts *ExecuteCommand) {
		opts.Env = env
	}
}

// WithExecuteTimeout sets the timeout for command execution.
//
// If the command doesn't complete within the timeout, it will be terminated.
//
// Example:
//
//	result, err := sandbox.Process.ExecuteCommand(ctx, "sleep 60",
//	    options.WithExecuteTimeout(5*time.Second),
//	)
func WithExecuteTimeout(timeout time.Duration) func(*ExecuteCommand) {
	return func(opts *ExecuteCommand) {
		opts.Timeout = &timeout
	}
}

type CodeRun struct {
	Params   *types.CodeRunParams // Code execution parameters
	Timeout  *time.Duration       // Execution timeout
	Language types.CodeLanguage   // Override the default language
}

// WithCodeRunParams sets the code execution parameters.
//
// Example:
//
//	result, err := sandbox.Process.CodeRun(ctx, code,
//	    options.WithCodeRunParams(types.CodeRunParams{Language: "python"}),
//	)
func WithCodeRunParams(params types.CodeRunParams) func(*CodeRun) {
	return func(opts *CodeRun) {
		opts.Params = &params
	}
}

func WithCodeRunLanguage(language types.CodeLanguage) func(*CodeRun) {
	return func(opts *CodeRun) {
		opts.Language = language
	}
}

func WithCodeRunTimeout(timeout time.Duration) func(*CodeRun) {
	return func(opts *CodeRun) {
		opts.Timeout = &timeout
	}
}

type PtySession struct {
	PtySize *types.PtySize    // Terminal dimensions (rows and columns)
	Env     map[string]string // Environment variables for the PTY session
}

// WithPtySize sets the PTY terminal dimensions.
//
// Example:
//
//	session, err := sandbox.Process.CreatePtySession(ctx, "my-session",
//	    options.WithPtySize(types.PtySize{Rows: 24, Cols: 80}),
//	)
func WithPtySize(size types.PtySize) func(*PtySession) {
	return func(opts *PtySession) {
		opts.PtySize = &size
	}
}

// WithPtyEnv sets environment variables for the PTY session.
//
// Example:
//
//	session, err := sandbox.Process.CreatePtySession(ctx, "my-session",
//	    options.WithPtyEnv(map[string]string{"TERM": "xterm-256color"}),
//	)
func WithPtyEnv(env map[string]string) func(*PtySession) {
	return func(opts *PtySession) {
		opts.Env = env
	}
}

type CreatePty struct {
	PtySize *types.PtySize    // Terminal dimensions (rows and columns)
	Env     map[string]string // Environment variables for the PTY session
}

// WithCreatePtySize sets the PTY terminal dimensions for CreatePty.
//
// Example:
//
//	handle, err := sandbox.Process.CreatePty(ctx, "my-pty",
//	    options.WithCreatePtySize(types.PtySize{Rows: 24, Cols: 80}),
//	)
func WithCreatePtySize(ptySize types.PtySize) func(*CreatePty) {
	return func(opts *CreatePty) {
		opts.PtySize = &ptySize
	}
}

// WithCreatePtyEnv sets environment variables for CreatePty.
//
// Example:
//
//	handle, err := sandbox.Process.CreatePty(ctx, "my-pty",
//	    options.WithCreatePtyEnv(map[string]string{"TERM": "xterm-256color"}),
//	)
func WithCreatePtyEnv(env map[string]string) func(*CreatePty) {
	return func(opts *CreatePty) {
		opts.Env = env
	}
}
