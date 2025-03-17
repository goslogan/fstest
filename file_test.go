package fs_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gotest.tools/v3/fs"
)

func TestNewDirWithOpsAndManifestEqual(t *testing.T) {
	var userOps []fs.PathOp
	if os.Geteuid() == 0 {
		userOps = append(userOps, fs.AsUser(1001, 1002))
	}

	ops := []fs.PathOp{
		fs.WithFile("file1", "contenta", fs.WithMode(0400)),
		fs.WithFile("file2", "", fs.WithBytes([]byte{0, 1, 2})),
		fs.WithFile("file5", "", userOps...),
		fs.WithSymlink("link1", "file1"),
		fs.WithDir("sub",
			fs.WithFiles(map[string]string{
				"file3": "contentb",
				"file4": "contentc",
			}),
			fs.WithMode(0705),
		),
	}

	dir := fs.NewDir(t, "test-all", ops...)
	defer dir.Remove()

	manifestOps := append(
		ops[:3],
		fs.WithSymlink("link1", dir.Join("file1")),
		ops[4],
	)
	assert.Assert(t, fs.Equal(dir.Path(), fs.Expected(t, manifestOps...)))
}

func TestNewFile(t *testing.T) {
	t.Run("with test name", func(t *testing.T) {
		tmpFile := fs.NewFile(t, t.Name())
		_, err := os.Stat(tmpFile.Path())
		assert.NilError(t, err)

		tmpFile.Remove()
		_, err = os.Stat(tmpFile.Path())
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run(`with \ in name`, func(t *testing.T) {
		tmpFile := fs.NewFile(t, `foo\thing`)
		_, err := os.Stat(tmpFile.Path())
		assert.NilError(t, err)

		tmpFile.Remove()
		_, err = os.Stat(tmpFile.Path())
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}

func TestNewFile_IntegrationWithCleanup(t *testing.T) {
	major, minor, err := GoVersion()
	assert.NilError(t, err)

	if major < 1 || (major == 1 && minor < 14) {
		t.Skip("skipping test because Go version is less than 1.14")
	}

	var tmpFile *fs.File
	t.Run("cleanup in subtest", func(t *testing.T) {
		tmpFile = fs.NewFile(t, t.Name())
		_, err := os.Stat(tmpFile.Path())
		assert.NilError(t, err)
	})

	t.Run("file has been removed", func(t *testing.T) {
		_, err := os.Stat(tmpFile.Path())
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}

func TestNewDir_IntegrationWithCleanup(t *testing.T) {

	major, minor, err := GoVersion()
	assert.NilError(t, err)

	if major < 1 || (major == 1 && minor < 14) {
		t.Skip("skipping test because Go version is less than 1.14")
	}

	var tmpFile *fs.Dir
	t.Run("cleanup in subtest", func(t *testing.T) {
		tmpFile = fs.NewDir(t, t.Name())
		_, err := os.Stat(tmpFile.Path())
		assert.NilError(t, err)
	})

	t.Run("dir has been removed", func(t *testing.T) {
		_, err := os.Stat(tmpFile.Path())
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}

func TestDirFromPath(t *testing.T) {
	tmpdir := t.TempDir()

	dir := fs.DirFromPath(t, tmpdir, fs.WithFile("newfile", ""))

	_, err := os.Stat(dir.Join("newfile"))
	assert.NilError(t, err)

	assert.Equal(t, dir.Path(), tmpdir)
	assert.Equal(t, dir.Join("newfile"), filepath.Join(tmpdir, "newfile"))

	dir.Remove()

	_, err = os.Stat(tmpdir)
	assert.Assert(t, errors.Is(err, os.ErrNotExist))
}

// Returns the go version as major, minor
func GoVersion() (int, int, error) {
	version := runtime.Version()
	if !strings.HasPrefix(version, "go") {
		return 0, 0, fmt.Errorf("runtime.Version() does appear to be a Go version string: %s", version)
	} else {
		version = strings.TrimPrefix(version, "go")
		parts := strings.Split(version, ".")
		if len(parts) < 2 {
			return 0, 0, fmt.Errorf("runtime.Version() does appear to be a Go version string: %s", version)
		}
		rMajor, err := strconv.ParseInt(parts[0], 10, 32)
		if err != nil {
			return 0, 0, fmt.Errorf("runtime.Version() does appear to be a Go version string: %s", version)
		}
		rMinor, err := strconv.ParseInt(parts[1], 10, 32)
		if err != nil {
			return 0, 0, fmt.Errorf("runtime.Version() does appear to be a Go version string: %s", version)
		}
		return int(rMajor), int(rMinor), nil
	}
}
