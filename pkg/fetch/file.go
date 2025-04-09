package fetch

import (
	"context"
	"strings"
)

func File(_ context.Context, src string) (string, error) {
	return strings.TrimPrefix(src, "file://"), nil
}
