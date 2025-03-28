/*
Package fs provides tools for creating temporary files, and testing the
contents and structure of a directory.
*/
package fs // import "gotest.tools/v3/fs"

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Path objects return their filesystem path. Path may be implemented by a
// real filesystem object (such as File and Dir) or by a type which updates
// entries in a Manifest.
type Path interface {
	Path() string
	Remove()
}

var (
	_ Path = &Dir{}
	_ Path = &File{}
)

// File is a temporary file on the filesystem
type File struct {
	path string
}

type helperT interface {
	Helper()
}

// NewFile creates a new file in a temporary directory using prefix as part of
// the filename. The PathOps are applied to the before returning the File.
//
// When used with Go 1.14+ the file will be automatically removed when the test
// ends, unless the TEST_NOCLEANUP env var is set to true.
func NewFile(t *testing.T, prefix string, ops ...PathOp) *File {
	tempfile, err := os.CreateTemp("", cleanPrefix(prefix)+"-")
	assert.Nil(t, err)

	file := &File{path: tempfile.Name()}
	t.Cleanup(file.Remove)

	assert.Nil(t, tempfile.Close())
	assert.Nil(t, applyPathOps(file, ops))
	return file
}

func cleanPrefix(prefix string) string {
	// windows requires both / and \ are replaced
	if runtime.GOOS == "windows" {
		prefix = strings.Replace(prefix, string(os.PathSeparator), "-", -1)
	}
	return strings.Replace(prefix, "/", "-", -1)
}

// Path returns the full path to the file
func (f *File) Path() string {
	return f.path
}

// Remove the file
func (f *File) Remove() {
	_ = os.Remove(f.path)
}

// Dir is a temporary directory
type Dir struct {
	path string
}

// NewDir returns a new temporary directory using prefix as part of the directory
// name. The PathOps are applied before returning the Dir.
//
// When used with Go 1.14+ the directory will be automatically removed when the test
// ends, unless the TEST_NOCLEANUP env var is set to true.
func NewDir(t *testing.T, prefix string, ops ...PathOp) *Dir {
	path, err := os.MkdirTemp("", cleanPrefix(prefix)+"-")
	assert.Nil(t, err)
	dir := &Dir{path: path}
	t.Cleanup(dir.Remove)

	assert.Nil(t, applyPathOps(dir, ops))
	return dir
}

// Path returns the full path to the directory
func (d *Dir) Path() string {
	return d.path
}

// Remove the directory
func (d *Dir) Remove() {
	_ = os.RemoveAll(d.path)
}

// Join returns a new path with this directory as the base of the path
func (d *Dir) Join(parts ...string) string {
	return filepath.Join(append([]string{d.Path()}, parts...)...)
}

// DirFromPath returns a Dir for a path that already exists. No directory is created.
// Unlike NewDir the directory will not be removed automatically when the test exits,
// it is the callers responsibly to remove the directory.
// DirFromPath can be used with Apply to modify an existing directory.
//
// If the path does not already exist, use NewDir instead.
func DirFromPath(t *testing.T, path string, ops ...PathOp) *Dir {

	dir := &Dir{path: path}
	assert.Nil(t, applyPathOps(dir, ops))
	return dir
}
