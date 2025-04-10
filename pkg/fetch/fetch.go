package fetch

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/Snakdy/container-build-engine/internal/decompress"
	"github.com/Snakdy/container-build-engine/pkg/fileutils"
	"github.com/go-logr/logr"
	"github.com/gosimple/hashdir"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

func Fetch(ctx context.Context, src, dst, checksum string) (string, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(6).Info("fetching file", "src", src, "dst", dst)

	uri, err := url.Parse(src)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}

	var out string
	if uri.Scheme == "https" {
		out, err = URL(ctx, uri)
	} else {
		out, err = File(ctx, uri)
	}
	if err != nil {
		return "", err
	}

	// also allow the checksum to be set in
	// the url query
	if checksum == "" {
		checksum = uri.Query().Get("checksum")
	}

	// verify the checksum of the file
	if checksum != "" {
		log.V(3).Info("verifying checksum", "checksum", checksum, "file", out)
		if err := checksumFile(out, checksum); err != nil {
			return "", err
		}
	}

	dontArchive := uri.Query().Get("archive") == "false"
	if dontArchive {
		log.V(3).Info("skipping unarchival process")
		out, err = fileutils.Copy(ctx, out, dst)
		if err != nil {
			return "", err
		}
		return out, nil
	}

	f, err := os.Open(out)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	out, err = decompress.Decompress(ctx, f, dst, out)
	if err != nil {
		return "", fmt.Errorf("decompressing: %w", err)
	}

	return out, nil
}

// checksumFile verifies the checksum of a file or
// directory.
func checksumFile(src, checksum string) error {
	// if the file is actually a directory then get a
	// checksum of the whole dir
	if info, err := os.Stat(src); err == nil && info.IsDir() {
		return checksumDir(src, checksum)
	}
	f, err := os.Open(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("generating sum: %w", err)
	}
	digest := hex.EncodeToString(h.Sum(nil))
	if digest != checksum {
		return fmt.Errorf("digests do not match: expected %s, got %s", checksum, digest)
	}
	return nil
}

func checksumDir(src, checksum string) error {
	digest, err := hashdir.Make(src, "sha256")
	if err != nil {
		return fmt.Errorf("hashing file: %w", err)
	}
	if digest != checksum {
		return fmt.Errorf("digests do not match: expected %s, got %s", checksum, digest)
	}
	return nil
}
