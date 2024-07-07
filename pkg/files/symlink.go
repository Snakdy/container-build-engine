package files

import (
	fullfs "chainguard.dev/apko/pkg/apk/fs"
)

func IsSymbolicLink(rootfs fullfs.FullFS, path string) (bool, error) {
	_, err := rootfs.Readlink(path)
	if err != nil {
		return false, nil
	}
	return true, nil
}
