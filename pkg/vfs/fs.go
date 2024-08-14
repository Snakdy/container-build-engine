package vfs

import (
	apkfs "chainguard.dev/apko/pkg/apk/fs"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	DefaultDirectoryPermissions = 0755
	DefaultFilePermissions      = 0644
)

// VFS is a partial implementation of the fs.FullFS interface that
// aims to trick a consumer into thinking it has privileged access
// to a rootfs, when it's actually just being dumped on the disk
// somewhere.
type VFS struct {
	path        string
	permissions map[string]fs.FileMode
	uid         map[string]int
	gid         map[string]int
}

func NewVFS(path string) *VFS {
	return &VFS{
		path:        path,
		permissions: map[string]fs.FileMode{},
		uid:         map[string]int{},
		gid:         map[string]int{},
	}
}

func (V *VFS) Path(path string) string {
	return Clean(filepath.Join(V.path, path))
}

func (V *VFS) Mkdir(path string, perm fs.FileMode) error {
	V.permissions[path] = perm
	return os.Mkdir(V.Path(path), DefaultDirectoryPermissions)
}

func (V *VFS) MkdirAll(path string, perm fs.FileMode) error {
	V.permissions[path] = perm
	return os.MkdirAll(V.Path(path), DefaultDirectoryPermissions)
}

func (V *VFS) Open(name string) (fs.File, error) {
	return os.Open(V.Path(name))
}

func (V *VFS) OpenReaderAt(name string) (apkfs.File, error) {
	//TODO implement me
	panic("implement me")
}

func (V *VFS) OpenFile(name string, flag int, perm fs.FileMode) (apkfs.File, error) {
	return os.OpenFile(V.Path(name), flag, perm)
}

func (V *VFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(V.Path(name))
}

func (V *VFS) WriteFile(name string, b []byte, mode fs.FileMode) error {
	V.permissions[name] = mode
	return os.WriteFile(V.Path(name), b, DefaultFilePermissions)
}

func (V *VFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(V.Path(name))
}

func (V *VFS) Mknod(path string, mode uint32, dev int) error {
	//TODO implement me
	panic("implement me")
}

func (V *VFS) Readnod(name string) (dev int, err error) {
	//TODO implement me
	panic("implement me")
}

func (V *VFS) Symlink(oldname, newname string) error {
	return os.Symlink(filepath.Join(V.path, oldname), filepath.Join(V.path, newname))
}

func (V *VFS) Link(oldname, newname string) error {
	return os.Link(filepath.Join(V.path, oldname), filepath.Join(V.path, newname))
}

func (V *VFS) Readlink(name string) (target string, err error) {
	return os.Readlink(V.Path(name))
}

func (V *VFS) Stat(path string) (fs.FileInfo, error) {
	return os.Stat(V.Path(path))
}

func (V *VFS) Lstat(path string) (fs.FileInfo, error) {
	return os.Lstat(V.Path(path))
}

func (V *VFS) Create(name string) (apkfs.File, error) {
	return os.Create(V.Path(name))
}

func (V *VFS) Remove(name string) error {
	return os.Remove(V.Path(name))
}

func (V *VFS) Chmod(path string, perm fs.FileMode) error {
	V.permissions[path] = perm
	//return os.Chmod(V.Path(path), perm)
	return nil
}

func (V *VFS) Chown(path string, uid int, gid int) error {
	V.uid[path] = uid
	V.gid[path] = gid
	//return os.Chown(V.Path(path), uid, gid)
	return nil
}

func (V *VFS) SetXattr(path string, attr string, data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (V *VFS) GetXattr(path string, attr string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (V *VFS) RemoveXattr(path string, attr string) error {
	//TODO implement me
	panic("implement me")
}

func (V *VFS) ListXattrs(path string) (map[string][]byte, error) {
	//TODO implement me
	panic("implement me")
}
