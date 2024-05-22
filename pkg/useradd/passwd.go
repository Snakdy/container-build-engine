package useradd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/chainguard-dev/go-apk/pkg/fs"
	"github.com/go-logr/logr"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const DefaultShell = "/bin/sh"

// NewUser adds an entry to the /etc/passwd file to create a new Linux
// user.
func NewUser(ctx context.Context, rootfs fs.FullFS, username string, uid int) error {
	log := logr.FromContextOrDiscard(ctx).WithValues("username", username, "uid", uid)
	log.V(1).Info("creating user")

	path := filepath.Join("/etc", "passwd")
	ok, err := containsUser(rootfs, path, username, uid)
	if err != nil {
		log.Error(err, "failed to check if user already exists")
		return err
	}
	if ok {
		log.V(1).Info("user already exists")
		return nil
	}

	if err := rootfs.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Error(err, "failed to establish directory structure")
		return err
	}
	var passwdFile string
	if data, err := rootfs.ReadFile(path); err != nil {
		if !os.IsNotExist(err) {
			log.Error(err, "failed to read passwd file")
			return err
		}
	} else {
		passwdFile = string(data)
	}
	passwd := []byte(fmt.Sprintf("%s:x:%d:0:Linux User,,,:%s:%s\n", username, uid, filepath.Join("/home", username), DefaultShell))
	log.V(5).Info("writing passwd file", "path", path, "content", string(passwd))
	passwdFile = passwdFile + string(passwd)
	if err := rootfs.WriteFile(path, []byte(passwdFile), 0644); err != nil {
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

// containsUser checks if a given /etc/passwd file contains a user.
// It checks for a match based the username or uid.
func containsUser(fs fs.FullFS, path, username string, uid int) (bool, error) {
	f, err := fs.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	br := bufio.NewScanner(f)
	for br.Scan() {
		s := br.Text()
		if strings.Contains(s, fmt.Sprintf("%s:x:", username)) {
			return true, nil
		}
		if strings.Contains(s, fmt.Sprintf(":x:%d:0:", uid)) {
			return true, nil
		}
	}
	if err := br.Err(); err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	return false, nil
}
