package useradd

import (
	"context"
	_ "embed"
	"github.com/chainguard-dev/go-apk/pkg/fs"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"testing"
)

//go:embed testdata/existing
var expected string

//go:embed testdata/single
var expectedEmpty string

func TestNewUser(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))

	setup := func(s string) fs.FullFS {
		rootfs := fs.NewMemFS()
		require.NoError(t, rootfs.MkdirAll("/etc", 0755))

		f, err := os.Open(s)
		require.NoError(t, err)

		path := filepath.Join("/etc", "passwd")
		out, err := rootfs.Create(path)
		require.NoError(t, err)

		_, _ = io.Copy(out, f)
		return rootfs
	}

	t.Run("empty", func(t *testing.T) {
		rootfs := fs.NewMemFS()

		err := NewUser(ctx, rootfs, "somebody", 1001)
		assert.NoError(t, err)

		data, err := rootfs.ReadFile(filepath.Join("/etc", "passwd"))
		require.NoError(t, err)
		assert.EqualValues(t, expectedEmpty, string(data))
	})

	t.Run("normal", func(t *testing.T) {
		rootfs := setup("./testdata/normal")

		err := NewUser(ctx, rootfs, "somebody", 1001)
		assert.NoError(t, err)

		data, err := rootfs.ReadFile(filepath.Join("/etc", "passwd"))
		require.NoError(t, err)
		assert.EqualValues(t, expectedEmpty, string(data))
	})

	t.Run("existing", func(t *testing.T) {
		rootfs := setup("./testdata/existing")

		err := NewUser(ctx, rootfs, "somebody", 1001)
		assert.NoError(t, err)

		data, err := rootfs.ReadFile(filepath.Join("/etc", "passwd"))
		require.NoError(t, err)
		assert.EqualValues(t, expectedEmpty, string(data))
	})
}

func TestNewUserDir(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))

	rootfs := fs.NewMemFS()

	err := NewUserDir(ctx, rootfs, "somebody", 1001)
	assert.NoError(t, err)

	_, err = rootfs.Stat(filepath.Join("/home", "somebody", ".local", "bin"))
	assert.NotErrorIs(t, err, os.ErrNotExist)
}
