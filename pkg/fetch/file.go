package fetch

import (
	"context"
	"net/url"
	"path/filepath"
)

func File(_ context.Context, src *url.URL) (string, error) {
	return filepath.Join(src.Hostname(), src.Path), nil
}
