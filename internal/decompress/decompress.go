package decompress

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/mholt/archives"
	"io"
	"os"
	"path/filepath"
)

func Decompress(ctx context.Context, input io.Reader, dst, src string) (string, error) {
	log := logr.FromContextOrDiscard(ctx)
	format, stream, err := archives.Identify(ctx, filepath.Base(src), input)
	if err != nil {
		if errors.Is(err, archives.NoMatch) {
			log.V(4).Info("skipping decompression, not an archive")
			return copyFile(ctx, src, dst)
		}
		return "", fmt.Errorf("identifying compression: %w", err)
	}
	var links []link
	err = format.(archives.Extractor).Extract(ctx, stream, func(ctx context.Context, info archives.FileInfo) error {
		out := filepath.Clean(filepath.Join(dst, info.NameInArchive))
		// if it's a directory, create it
		if info.IsDir() {
			if err := os.MkdirAll(out, info.Mode()); err != nil {
				return fmt.Errorf("creating directory: %w", err)
			}
			return nil
		}
		// if it's a link, take note
		if info.LinkTarget != "" {
			links = append(links, link{
				Source: filepath.Join(filepath.Dir(out), info.LinkTarget),
				Target: out,
			})
			return nil
		}

		log.V(9).Info("extracting", "src", info.NameInArchive, "dst", out)

		// create the file
		if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
			return fmt.Errorf("ensuring directory: %w", err)
		}
		f, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
		if err != nil {
			return fmt.Errorf("opening file for writing: %w", err)
		}
		defer f.Close()
		in, err := info.Open()
		if err != nil {
			return fmt.Errorf("opening file for reading: %w", err)
		}
		defer in.Close()

		_, err = io.Copy(f, in)
		if err != nil {
			return fmt.Errorf("copying file: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("extracting archive: %w", err)
	}
	for _, l := range links {
		log.V(9).Info("linking", "src", l.Source, "dst", l.Target)
		if err := os.Symlink(l.Source, l.Target); err != nil {
			return "", fmt.Errorf("linking: %w", err)
		}
	}
	return dst, nil
}

func copyFile(ctx context.Context, src, dst string) (string, error) {
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
