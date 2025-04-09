package fetch

import (
	"context"
	"fmt"
	"github.com/carlmjohnson/requests"
	"github.com/go-logr/logr"
	"os"
	"path"
	"path/filepath"
)

func URL(ctx context.Context, src string) (string, error) {
	log := logr.FromContextOrDiscard(ctx)

	tmp, err := os.MkdirTemp("", "temp-download-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}
	dst := filepath.Join(tmp, path.Base(src))

	log.V(6).Info("downloading file", "src", src, "dst", dst)

	err = requests.URL(src).ToFile(dst).Fetch(ctx)
	if err != nil {
		_ = os.Remove(dst)
		return "", fmt.Errorf("downloading file: %w", err)
	}
	return dst, nil
}
