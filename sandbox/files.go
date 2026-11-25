package sandbox

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/digitalocean/godo"
)

// Files reads and writes the files of one sandbox. Get it from Sandbox.Files.
//
// Paths are inside the sandbox workspace; relative paths are resolved from the
// workspace root. Write, WriteArchive, Read and ReadArchive move up to
// godo.SandboxFileMaxBytes (50 MiB) in one call.
type Files struct {
	client    *Client
	sandboxID string
}

// FileType is the type of a FileEntry.
type FileType string

const (
	// FileTypeUnspecified is an entry whose type the API did not report.
	FileTypeUnspecified FileType = "unspecified"
	// FileTypeFile is a regular file.
	FileTypeFile FileType = "file"
	// FileTypeDirectory is a directory.
	FileTypeDirectory FileType = "directory"
	// FileTypeSymlink is a symbolic link. See FileEntry.SymlinkTarget.
	FileTypeSymlink FileType = "symlink"
)

// FileEntry is a file, directory or symlink in a sandbox.
type FileEntry struct {
	// Name is the base name of the entry.
	Name string
	Type FileType
	// Path is the absolute path of the entry in the sandbox.
	Path string
	// Size is the size in bytes.
	Size int64
	// Mode is the Unix file mode bits, such as 0644.
	Mode int64
	// Permissions are the permissions in ls form, such as "-rw-r--r--".
	Permissions string
	Owner       string
	Group       string
	// ModifiedTime is when the entry was last modified.
	ModifiedTime time.Time
	// SymlinkTarget is where a symlink points. Empty for other types.
	SymlinkTarget string
}

// IsDir reports whether the entry is a directory.
func (e *FileEntry) IsDir() bool { return e.Type == FileTypeDirectory }

func newFileEntry(e godo.SandboxFileEntry) *FileEntry {
	t := FileTypeUnspecified
	switch e.Type {
	case godo.SandboxFileEntryTypeFile:
		t = FileTypeFile
	case godo.SandboxFileEntryTypeDirectory:
		t = FileTypeDirectory
	case godo.SandboxFileEntryTypeSymlink:
		t = FileTypeSymlink
	}
	return &FileEntry{
		Name: e.Name, Type: t, Path: e.Path, Size: e.Size, Mode: e.Mode,
		Permissions: e.Permissions, Owner: e.Owner, Group: e.Group,
		ModifiedTime: e.ModifiedTime.Time, SymlinkTarget: e.SymlinkTarget,
	}
}

// WriteInfo is the result of Write and WriteArchive.
type WriteInfo struct {
	// Path is the absolute path written to.
	Path string
	// BytesWritten is the number of bytes written.
	BytesWritten int64
}

// ListFilesOptions are the options for List. Every field is optional.
type ListFilesOptions struct {
	// Depth is how many levels to list. 0 or 1 lists the directory's own
	// entries, 2 also lists the entries of its subdirectories, and so on.
	Depth int
}

func checkPath(name, path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("sandbox: " + name + " must not be empty")
	}
	return nil
}

// Write writes data to a file, replacing it if it exists and creating any
// missing parent directories. The SHA-256 of data is sent, and the sandbox
// checks it after the write.
func (f *Files) Write(ctx context.Context, path string, data []byte) (*WriteInfo, error) {
	return f.upload(ctx, path, data, false)
}

// WriteArchive extracts a tar archive into the directory dir.
func (f *Files) WriteArchive(ctx context.Context, dir string, tar []byte) (*WriteInfo, error) {
	return f.upload(ctx, dir, tar, true)
}

func (f *Files) upload(ctx context.Context, path string, data []byte, archive bool) (*WriteInfo, error) {
	if err := checkPath("path", path); err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	out, _, err := f.client.API.Sandboxes.UploadFile(ctx, f.sandboxID, &godo.SandboxFileUploadRequest{
		Path:      path,
		Body:      bytes.NewReader(data),
		IsArchive: archive,
		SHA256:    hex.EncodeToString(sum[:]),
	})
	if err != nil {
		return nil, err
	}
	return &WriteInfo{Path: out.Path, BytesWritten: out.BytesWritten}, nil
}

// Read reads a file. The content is checked against the SHA-256 the sandbox
// sends with it. Reading a directory fails: use ReadArchive.
func (f *Files) Read(ctx context.Context, path string) ([]byte, error) {
	return f.download(ctx, path, false)
}

// ReadArchive reads a file or a directory as a tar archive.
func (f *Files) ReadArchive(ctx context.Context, path string) ([]byte, error) {
	return f.download(ctx, path, true)
}

func (f *Files) download(ctx context.Context, path string, archive bool) ([]byte, error) {
	if err := checkPath("path", path); err != nil {
		return nil, err
	}
	data, _, err := f.client.API.Sandboxes.DownloadFile(ctx, f.sandboxID, &godo.SandboxFileDownloadRequest{
		Path: path, AsArchive: archive,
	})
	return data, err
}

// List lists the entries of a directory. Symlinks are listed, not followed. A
// nil opts lists one level.
func (f *Files) List(ctx context.Context, path string, opts *ListFilesOptions) ([]FileEntry, error) {
	if err := checkPath("path", path); err != nil {
		return nil, err
	}
	depth := 0
	if opts != nil {
		if opts.Depth < 0 {
			return nil, errors.New("sandbox: Depth must not be negative")
		}
		depth = opts.Depth
	}
	entries, _, err := f.client.API.Sandboxes.ListFiles(ctx, f.sandboxID, path, depth)
	if err != nil {
		return nil, err
	}
	out := make([]FileEntry, len(entries))
	for i := range entries {
		out[i] = *newFileEntry(entries[i])
	}
	return out, nil
}

// GetInfo describes a file, directory or symlink. A missing path is a 404
// (IsNotFound); Exists tests for a path without an error.
func (f *Files) GetInfo(ctx context.Context, path string) (*FileEntry, error) {
	if err := checkPath("path", path); err != nil {
		return nil, err
	}
	entry, _, err := f.client.API.Sandboxes.GetFileInfo(ctx, f.sandboxID, path)
	if err != nil {
		return nil, err
	}
	return newFileEntry(*entry), nil
}

// Exists reports whether a file, directory or symlink exists at path.
func (f *Files) Exists(ctx context.Context, path string) (bool, error) {
	if err := checkPath("path", path); err != nil {
		return false, err
	}
	ok, _, err := f.client.API.Sandboxes.FileExists(ctx, f.sandboxID, path)
	return ok, err
}

// MakeDir creates a directory with any missing parents. created is false when
// the directory already existed; entry describes it in both cases.
func (f *Files) MakeDir(ctx context.Context, path string) (created bool, entry *FileEntry, err error) {
	if err := checkPath("path", path); err != nil {
		return false, nil, err
	}
	out, _, err := f.client.API.Sandboxes.MakeDirectory(ctx, f.sandboxID, path)
	if err != nil {
		return false, nil, err
	}
	return out.Created, newFileEntry(out.Entry), nil
}

// Move moves or renames a file, directory or symlink and returns the entry at
// its new path. Missing parents of destination are created, and a file
// already at destination is replaced.
func (f *Files) Move(ctx context.Context, source, destination string) (*FileEntry, error) {
	if err := checkPath("source", source); err != nil {
		return nil, err
	}
	if err := checkPath("destination", destination); err != nil {
		return nil, err
	}
	entry, _, err := f.client.API.Sandboxes.MoveFile(ctx, f.sandboxID, source, destination)
	if err != nil {
		return nil, err
	}
	return newFileEntry(*entry), nil
}

// Remove removes a file, or a directory with everything in it. Removing a
// path that does not exist succeeds.
func (f *Files) Remove(ctx context.Context, path string) error {
	if err := checkPath("path", path); err != nil {
		return err
	}
	_, err := f.client.API.Sandboxes.RemoveFile(ctx, f.sandboxID, path)
	return err
}
