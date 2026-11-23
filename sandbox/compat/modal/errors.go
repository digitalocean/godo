// Copied from github.com/modal-labs/modal-client/go v0.11.2 (Apache-2.0); see NOTICE.

package modal

type FunctionTimeoutError struct {
	Exception string

	// FunctionCall.Get wait timeouts stay a FunctionTimeoutError for backward compatibility and
	// unwrap to TimeoutError.
	waitTimeout bool
}

func (e FunctionTimeoutError) Error() string {
	return "FunctionTimeoutError: " + e.Exception
}

func (e FunctionTimeoutError) Unwrap() error {
	if e.waitTimeout {
		return TimeoutError{e.Exception}
	}
	return nil
}

// RemoteError represents an error on the Modal server, or a Python exception.
type RemoteError struct {
	Exception string
}

func (e RemoteError) Error() string {
	return "RemoteError: " + e.Exception
}

// InternalFailure is a retryable internal error from Modal.
type InternalFailure struct {
	Exception string
}

func (e InternalFailure) Error() string {
	return "InternalFailure: " + e.Exception
}

// ExecutionError is returned when something unexpected happened during runtime.
type ExecutionError struct {
	Exception string
}

func (e ExecutionError) Error() string {
	return "ExecutionError: " + e.Exception
}

// NotFoundError is returned when a resource is not found.
type NotFoundError struct {
	Exception string
}

func (e NotFoundError) Error() string {
	return "NotFoundError: " + e.Exception
}

// AlreadyExistsError is returned when a resource already exists.
type AlreadyExistsError struct {
	Exception string
}

func (e AlreadyExistsError) Error() string {
	return "AlreadyExistsError: " + e.Exception
}

// ConflictError is returned when the current state of a resource conflicts
// with the requested operation.
type ConflictError struct {
	Exception string
}

func (e ConflictError) Error() string {
	return "ConflictError: " + e.Exception
}

// InvalidError represents an invalid request or operation.
type InvalidError struct {
	Exception string
}

func (e InvalidError) Error() string {
	return "InvalidError: " + e.Exception
}

// QueueEmptyError is returned when an operation is attempted on an empty Queue.
type QueueEmptyError struct {
	Exception string
}

func (e QueueEmptyError) Error() string {
	return "QueueEmptyError: " + e.Exception
}

// QueueFullError is returned when an operation is attempted on a full Queue.
type QueueFullError struct {
	Exception string
}

func (e QueueFullError) Error() string {
	return "QueueFullError: " + e.Exception
}

// ResourceExhaustedError is returned when a server-side resource has been
// exhausted, such as capacity, a quota, or a rate limit.
type ResourceExhaustedError struct {
	Exception string
}

func (e ResourceExhaustedError) Error() string {
	return "ResourceExhaustedError: " + e.Exception
}

// SandboxFilesystemError is returned for unexpected or unclassified sandbox filesystem errors.
type SandboxFilesystemError struct {
	Exception string
}

func (e SandboxFilesystemError) Error() string {
	return "SandboxFilesystemError: " + e.Exception
}

// SandboxFilesystemNotFoundError is returned when a path does not exist in the sandbox filesystem.
type SandboxFilesystemNotFoundError struct{ Exception string }

func (e SandboxFilesystemNotFoundError) Error() string {
	return "SandboxFilesystemNotFoundError: " + e.Exception
}

// SandboxFilesystemNotADirectoryError is returned when a directory operation targets a non-directory path.
type SandboxFilesystemNotADirectoryError struct{ Exception string }

func (e SandboxFilesystemNotADirectoryError) Error() string {
	return "SandboxFilesystemNotADirectoryError: " + e.Exception
}

// SandboxFilesystemIsADirectoryError is returned when a file operation targets a directory.
type SandboxFilesystemIsADirectoryError struct{ Exception string }

func (e SandboxFilesystemIsADirectoryError) Error() string {
	return "SandboxFilesystemIsADirectoryError: " + e.Exception
}

// SandboxFilesystemPermissionError is returned when the sandbox filesystem denies access.
type SandboxFilesystemPermissionError struct{ Exception string }

func (e SandboxFilesystemPermissionError) Error() string {
	return "SandboxFilesystemPermissionError: " + e.Exception
}

// SandboxFilesystemPathAlreadyExistsError is returned when a create operation targets an existing path.
type SandboxFilesystemPathAlreadyExistsError struct{ Exception string }

func (e SandboxFilesystemPathAlreadyExistsError) Error() string {
	return "SandboxFilesystemPathAlreadyExistsError: " + e.Exception
}

// SandboxFilesystemDirectoryNotEmptyError is returned when a non-recursive remove operation targets a non-empty directory.
type SandboxFilesystemDirectoryNotEmptyError struct{ Exception string }

func (e SandboxFilesystemDirectoryNotEmptyError) Error() string {
	return "SandboxFilesystemDirectoryNotEmptyError: " + e.Exception
}

// SandboxFilesystemFileTooLargeError is returned when a file exceeds the allowed read size.
type SandboxFilesystemFileTooLargeError struct{ Exception string }

func (e SandboxFilesystemFileTooLargeError) Error() string {
	return "SandboxFilesystemFileTooLargeError: " + e.Exception
}

// SnapshotCreationError is returned when a snapshot operation is done and failed.
// Polling again will not produce an Image; filesystem state is gone.
type SnapshotCreationError struct {
	Exception string
}

func (e SnapshotCreationError) Error() string {
	return "SnapshotCreationError: " + e.Exception
}

// SandboxTimeoutError is returned when Sandbox operations exceed the allowed time limit.
type SandboxTimeoutError struct {
	Exception string
}

func (e SandboxTimeoutError) Error() string {
	return "SandboxTimeoutError: " + e.Exception
}

// TimeoutError is returned when an operation exceeds the allowed time limit.
type TimeoutError struct {
	Exception string
}

func (e TimeoutError) Error() string {
	return "TimeoutError: " + e.Exception
}

// ClientClosedError is returned when Sandbox operations exceed the allowed time limit.
type ClientClosedError struct {
	Exception string
}

func (e ClientClosedError) Error() string {
	return "ClientClosedError: " + e.Exception
}

// ExecTimeoutError is returned when a container exec exceeds its execution duration limit.
type ExecTimeoutError struct {
	Exception string
}

func (e ExecTimeoutError) Error() string {
	return "ExecTimeoutError: " + e.Exception
}
