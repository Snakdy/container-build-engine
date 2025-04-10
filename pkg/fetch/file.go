package fetch

import (
	"context"
	"net/url"
	"strings"
)

func File(_ context.Context, src *url.URL) (string, error) {
	return strings.TrimPrefix(src.String(), "file://"), nil
}
