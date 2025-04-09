package fetch

import (
	"context"
	"fmt"
	"github.com/Snakdy/container-build-engine/internal/decompress"
	"github.com/go-logr/logr"
	"github.com/gosimple/hashdir"
	"os"
	"strings"
)

func Fetch(ctx context.Context, src, dst, checksum string) (string, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(6).Info("fetching file", "src", src, "dst", dst)

	var out string
	var err error
	if strings.HasPrefix(src, "https://") {
		out, err = URL(ctx, src)
	} else {
		out, err = File(ctx, src)
	}
	if err != nil {
		return "", err
	}

	f, err := os.Open(out)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}

	// verify the checksum of the file
	if checksum != "" {
		if err := checksumFile(out, checksum); err != nil {
			return "", err
		}
	}

	out, err = decompress.Decompress(ctx, f, dst, out)
	if err != nil {
		return "", fmt.Errorf("decompressing: %w", err)
	}

	return out, nil
}

func checksumFile(src, checksum string) error {
	digest, err := hashdir.Make(src, "sha256")
	if err != nil {
		return fmt.Errorf("hashing file: %w", err)
	}
	if digest != checksum {
		return fmt.Errorf("digests do not match: expected %s, got %s", checksum, digest)
	}
	return nil
}
