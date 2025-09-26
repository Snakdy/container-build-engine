package files

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"chainguard.dev/apko/pkg/apk/fs"
	"github.com/go-logr/logr"
)

func CopyDirectory(ctx context.Context, srcDir, dest string, destFS fs.FullFS) error {
	log := logr.FromContextOrDiscard(ctx).WithValues("src", srcDir, "dst", dest)
	log.V(6).Info("copying directory")
	info, err := os.Stat(srcDir)
	if err != nil {
		return err
	}
	// check if the directory is actually a directory
	if !info.IsDir() {
		log.V(6).Info("directory is actually a file")
		return Copy(ctx, srcDir, dest, destFS)
	}
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(srcDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destFS, destPath, 0755); err != nil {
				return err
			}
			if err := CopyDirectory(ctx, sourcePath, destPath, destFS); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath, destFS); err != nil {
				return err
			}
		default:
			if err := Copy(ctx, sourcePath, destPath, destFS); err != nil {
				return err
			}
		}

		fInfo, err := entry.Info()
		if err != nil {
			return err
		}

		isSymlink := fInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := destFS.Chmod(destPath, fInfo.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func Copy(ctx context.Context, srcFile, dstFile string, dstFS fs.FullFS) error {
	log := logr.FromContextOrDiscard(ctx).WithValues("src", srcFile, "dst", dstFile)
	log.V(6).Info("copying file")
	if err := CreateIfNotExists(dstFS, filepath.Dir(dstFile), 0755); err != nil {
		return fmt.Errorf("ensuring heirarchy: %w", err)
	}
	out, err := dstFS.Create(dstFile)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	// get source file info
	info, err := in.Stat()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	// copy permissions from old to new
	if err := dstFS.Chmod(dstFile, info.Mode()); err != nil {
		return fmt.Errorf("chmoding file: %w", err)
	}
	if err := dstFS.Chown(dstFile, 1001, 0); err != nil {
		return fmt.Errorf("chowning file: %w", err)
	}

	return nil
}

func Exists(fs fs.FullFS, filePath string) bool {
	if _, err := fs.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateIfNotExists(fs fs.FullFS, dir string, perm os.FileMode) error {
	if Exists(fs, dir) {
		return nil
	}

	if err := fs.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("creating directory: '%s': '%s'", dir, err.Error())
	}

	return nil
}

func CopySymLink(source, dest string, dstFS fs.FullFS) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return dstFS.Symlink(link, dest)
}
