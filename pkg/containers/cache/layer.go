package cache

import (
	"fmt"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func create(path string, h v1.Hash) (io.WriteCloser, error) {
	if err := os.MkdirAll(path, 0700); err != nil {
		return nil, err
	}
	return os.Create(cachepath(path, h))
}

func cachepath(path string, h v1.Hash) string {
	file := h.String()
	if runtime.GOOS == "windows" {
		file = fmt.Sprintf("%s-%s", h.Algorithm, h.Hex)
	}
	return filepath.Join(path, file)
}
