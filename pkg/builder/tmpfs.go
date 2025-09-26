package builder

import (
	"context"
	"os"

	"chainguard.dev/apko/pkg/apk/fs"
)

// NewDirFS is a utility function to create a filesystem in a temporary directory
func NewDirFS(ctx context.Context) (fs.FullFS, error) {
	tmpFs, err := os.MkdirTemp("", "container-build-engine-fs-*")
	if err != nil {
		return nil, err
	}
	return fs.DirFS(ctx, tmpFs), nil
}
