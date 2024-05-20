package files

import (
	fullfs "github.com/chainguard-dev/go-apk/pkg/fs"
)

func IsSymbolicLink(rootfs fullfs.FullFS, path string) (bool, error) {
	_, err := rootfs.Readlink(path)
	if err != nil {
		return false, nil
	}
	return true, nil
}
