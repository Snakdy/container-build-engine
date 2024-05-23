package useradd

import (
	"context"
	"fmt"
	"github.com/chainguard-dev/go-apk/pkg/fs"
	"github.com/go-logr/logr"
	"path/filepath"
)

const DefaultShell = "/bin/sh"

// NewUser adds an entry to the /etc/passwd file to create a new Linux
// user.
func NewUser(ctx context.Context, rootfs fs.FullFS, username string, uid int) error {
	log := logr.FromContextOrDiscard(ctx).WithValues("username", username, "uid", uid)
	log.V(1).Info("creating user")

	path := filepath.Join("/etc", "passwd")
	if err := rootfs.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Error(err, "failed to establish directory structure")
		return err
	}
	passwd := []byte(fmt.Sprintf("%s:x:%d:0:Linux User,,,:%s:%s\n", username, uid, filepath.Join("/home", username), DefaultShell))
	log.V(5).Info("writing passwd file", "path", path, "content", string(passwd))
	if err := rootfs.WriteFile(path, passwd, 0644); err != nil {
		log.Error(err, "failed to write to passwd file")
		return err
	}

	// create the home directory.
	// hopefully the permission bits are correct - https://superuser.com/a/165465
	log.V(1).Info("creating user home directory")
	if err := rootfs.MkdirAll(filepath.Join("/home", username, ".local", "bin"), 0775); err != nil {
		log.Error(err, "failed to create home directory")
		return err
	}
	if err := rootfs.Chown(filepath.Join("/home", username), uid, 0); err != nil {
		log.Error(err, "failed to set home directory ownership")
		return err
	}

	return nil
}
