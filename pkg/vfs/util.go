package vfs

import (
	"path"
	"path/filepath"
	"strings"
)

func Clean(s string) string {
	return filepath.FromSlash(path.Clean("/" + strings.Trim(s, "/")))
}
