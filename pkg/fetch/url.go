package fetch

import (
	"context"
	"fmt"
	"github.com/carlmjohnson/requests"
	"github.com/go-logr/logr"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func URL(ctx context.Context, src *url.URL) (string, error) {
	log := logr.FromContextOrDiscard(ctx)

	tmp, err := os.MkdirTemp("", "temp-download-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}
	dst := filepath.Join(tmp, path.Base(src.Path))

	log.V(6).Info("downloading file", "src", src, "dst", dst)

	err = requests.URL(fmt.Sprintf("%s://%s/%s", src.Scheme, src.Host, strings.TrimPrefix(src.Path, "/"))).ToFile(dst).Fetch(ctx)
	if err != nil {
		_ = os.Remove(dst)
		return "", fmt.Errorf("downloading file: %w", err)
	}
	return dst, nil
}
