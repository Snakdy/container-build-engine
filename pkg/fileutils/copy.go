package fileutils

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"io"
	"os"
	"path/filepath"
)

func Copy(ctx context.Context, src, dst string) (string, error) {
	log := logr.FromContextOrDiscard(ctx)
	out := filepath.Join(dst, filepath.Base(src))
	log.V(6).Info("copying file", "src", src, "dst", out)

	in, err := os.Open(filepath.Clean(src))
	if err != nil {
		return "", fmt.Errorf("opening file for reading: %w", err)
	}
	defer in.Close()

	// stat the file so that we can
	// set permissions properly
	info, err := in.Stat()
	if err != nil {
		return "", fmt.Errorf("stating file for reading: %w", err)
	}

	f, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return "", fmt.Errorf("opening file for writing: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, in)
	if err != nil {
		return "", fmt.Errorf("copying file: %w", err)
	}

	return out, nil
}
