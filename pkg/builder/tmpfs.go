package builder

import (
	"chainguard.dev/apko/pkg/apk/fs"
	"os"
)

// NewDirFS is a utility function to create a filesystem in a temporary directory
func NewDirFS() (fs.FullFS, error) {
	tmpFs, err := os.MkdirTemp("", "container-build-engine-fs-*")
	if err != nil {
		return nil, err
	}
	return fs.DirFS(tmpFs), nil
}
