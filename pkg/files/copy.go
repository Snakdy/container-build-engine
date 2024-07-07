package files

import (
	"chainguard.dev/apko/pkg/apk/fs"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func CopyDirectory(srcDir, dest string, destFS fs.FullFS) error {
	info, err := os.Stat(srcDir)
	if err != nil {
		return err
	}
	// check if the directory is actually a directory
	if !info.IsDir() {
		return Copy(srcDir, dest, destFS)
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
			if err := CopyDirectory(sourcePath, destPath, destFS); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath, destFS); err != nil {
				return err
			}
		default:
			if err := Copy(sourcePath, destPath, destFS); err != nil {
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

func Copy(srcFile, dstFile string, dstFS fs.FullFS) error {
	if err := CreateIfNotExists(dstFS, filepath.Dir(dstFile), 0755); err != nil {
		return err
	}
	out, err := dstFS.Create(dstFile)
	if err != nil {
		return err
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
		return err
	}
	if err := dstFS.Chown(dstFile, 1001, 0); err != nil {
		return err
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
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
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
