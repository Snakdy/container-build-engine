package useradd

import (
	"chainguard.dev/apko/pkg/apk/fs"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"path/filepath"
	"strings"
)

const (
	DefaultShell = "/bin/sh"
	RootUser     = "root"
)

// NewUser adds an entry to the /etc/passwd file to create a new Linux
// user. This must be run after regular filesystem mutations (i.e. just before the layer is appended)
func NewUser(ctx context.Context, rootfs fs.FullFS, username, shell string, uid int) error {
	log := logr.FromContextOrDiscard(ctx).WithValues("username", username, "uid", uid, "shell", shell)
	log.V(1).Info("creating user")

	path := filepath.Join("/etc", "passwd")
	if err := rootfs.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Error(err, "failed to establish directory structure")
		return err
	}
	// allow the user to provide a custom shell
	if strings.TrimSpace(shell) == "" {
		shell = DefaultShell
	}
	passwd := []byte(fmt.Sprintf("%s:x:%d:0:Linux User,,,:%s:%s\nroot:x:0:0:root:/root:%s\n", username, uid, filepath.Join("/home", username), shell, shell))
	log.V(5).Info("writing passwd file", "path", path, "content", string(passwd))
	if err := rootfs.WriteFile(path, passwd, 0644); err != nil {
		log.Error(err, "failed to write to passwd file")
		return err
	}

	return nil
}

// NewUserDir creates the filesystem for a new Linux
// user.
func NewUserDir(ctx context.Context, rootfs fs.FullFS, username string, uid int) error {
	log := logr.FromContextOrDiscard(ctx).WithValues("username", username, "uid", uid)
	log.V(1).Info("creating user filesystem")

	homedir := filepath.Join("/home", username)
	if username == RootUser {
		homedir = "/root"
	}

	// create the home directory.
	// hopefully the permission bits are correct - https://superuser.com/a/165465
	log.V(1).Info("creating user home directory")
	if err := rootfs.MkdirAll(filepath.Join(homedir, ".local", "bin"), 0775); err != nil {
		log.Error(err, "failed to create home directory")
		return err
	}
	if err := rootfs.Chown(homedir, uid, 0); err != nil {
		log.Error(err, "failed to set home directory ownership")
		return err
	}

	return nil
}
