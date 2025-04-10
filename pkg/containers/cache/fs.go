package cache

import (
	"errors"
	"fmt"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/cache"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"io"
	"os"
	"path/filepath"
)

// fscache is a modification of the Crane cache
// that allows explicit specification of the layer
// key.
//
// It also writes the layer straight away rather
// than doing it lazily.
type fscache struct {
	path string
}

func NewFilesystemCache(path string) Cache {
	return &fscache{
		path,
	}
}

func (fs *fscache) Put(key v1.Hash, layer v1.Layer, compressed bool) (v1.Layer, error) {
	var rc io.ReadCloser
	var err error
	if compressed {
		rc, err = layer.Compressed()
	} else {
		rc, err = layer.Uncompressed()
	}
	if err != nil {
		return nil, fmt.Errorf("reading layer: %w", err)
	}
	// create the file
	w, err := create(fs.path, key)
	if err != nil {
		return nil, fmt.Errorf("preparing fs: %w", err)
	}
	defer w.Close()
	// copy the layer into the file
	_, err = io.Copy(w, rc)
	if err != nil {
		return nil, fmt.Errorf("writing layer: %w", err)
	}
	return layer, nil
}

func (fs *fscache) Get(hash v1.Hash) (v1.Layer, error) {
	path := cachepath(fs.path, hash)
	// try to read the layer from a file
	l, err := tarball.LayerFromFile(path)
	if os.IsNotExist(err) {
		return nil, cache.ErrNotFound
	}
	// if it's somehow corrupt, delete it
	// so we can fix it next time around
	if errors.Is(err, io.ErrUnexpectedEOF) {
		if err := fs.Delete(hash); err != nil {
			return nil, err
		}
		return nil, cache.ErrNotFound
	}
	return l, nil
}

func (fs *fscache) Delete(hash v1.Hash) error {
	if err := os.Remove(cachepath(fs.path, hash)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cache.ErrNotFound
		}
		return err
	}
	return nil
}

const layerDir = "cbe"

// Dir attempts to find a directory that we can store the
// cached layers in.
// It will try the XDG_CACHE_HOME first, followed by ~/.cache
// and finally TMPDIR or /tmp
func Dir() string {
	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = os.TempDir()
		}
		cacheHome = filepath.Join(home, ".cache")
	}
	cacheHome = filepath.Join(cacheHome, layerDir)
	_ = os.MkdirAll(cacheHome, 0750)
	return cacheHome
}
