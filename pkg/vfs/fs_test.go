package vfs

import "chainguard.dev/apko/pkg/apk/fs"

var _ fs.FullFS = &VFS{}
