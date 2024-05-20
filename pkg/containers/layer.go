package containers

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"github.com/Snakdy/container-build-engine/pkg/files"
	fullfs "github.com/chainguard-dev/go-apk/pkg/fs"
	"github.com/go-logr/logr"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

func NewLayer(ctx context.Context, fs fullfs.FullFS, platform *v1.Platform) (v1.Layer, error) {
	layerBuf, err := tarDir(ctx, fs, platform)
	if err != nil {
		return nil, fmt.Errorf("tarring data: %w", err)
	}
	layerBytes := layerBuf.Bytes()
	return tarball.LayerFromOpener(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBuffer(layerBytes)), nil
	}, tarball.WithCompressedCaching, tarball.WithMediaType(types.OCILayer))
}

func tarDir(ctx context.Context, fs fullfs.FullFS, platform *v1.Platform) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	if err := walkRecursive(ctx, fs, tw, "/", v1.Time{}, platform); err != nil {
		return nil, err
	}
	return buf, nil
}

// walkRecursive performs a filepath.Walk of the given root directory adding it
// to the provided tar.Writer with root -> chroot.  All symlinks are dereferenced,
// which is what leads to recursion when we encounter a directory symlink.
func walkRecursive(ctx context.Context, rootfs fullfs.FullFS, tw *tar.Writer, root string, creationTime v1.Time, platform *v1.Platform) error {
	log := logr.FromContextOrDiscard(ctx).WithValues("root", root)
	log.V(1).Info("walking filesystem")
	dirs, err := fs.ReadDir(rootfs, root)
	if err != nil {
		return fmt.Errorf("fs.ReadDir(%q): %w", root, err)
	}
	for _, d := range dirs {
		hostPath := filepath.Join(root, d.Name())
		log.V(1).Info("checking file", "path", hostPath, "dir", d.IsDir())
		if hostPath == root || hostPath == "/" {
			continue
		}

		// hacky method of setting the uid...
		uid := 0
		if hostPath == "/home/somebody" || strings.HasPrefix(hostPath, "/home/somebody") {
			log.V(4).Info("adding user owned file")
			uid = 1001
		}

		// create directory shells
		if d.IsDir() {
			log.V(4).Info("adding directory to tar", "dir", hostPath)
			header := &tar.Header{
				Name:     hostPath,
				Typeflag: tar.TypeDir,
				Mode:     0775,
				ModTime:  creationTime.Time,
				Uid:      uid,
			}
			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("tar.Writer.WriteHeader(%q): %w", hostPath, err)
			}
		}

		evalPath := hostPath
		ok, err := files.IsSymbolicLink(rootfs, hostPath)
		if err != nil {
			return fmt.Errorf("fileutil.IsSymbolicLink(%q): %w", hostPath, err)
		}
		if ok {
			log.V(5).Info("expanding symbolic link")
			evalPath, err = rootfs.Readlink(hostPath)
			if err != nil {
				return fmt.Errorf("fs.Readlink(%q): %w", hostPath, err)
			}
			log.V(4).Info("adding symbolic link to tar")
			header := &tar.Header{
				Name:     hostPath,
				Typeflag: tar.TypeSymlink,
				Linkname: evalPath,
				ModTime:  creationTime.Time,
				Uid:      uid,
			}
			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("tar.Writer.WriteHeader(%q): %w", hostPath, err)
			}
			continue
		}

		// Chase symlinks.
		info, err := fs.Stat(rootfs, evalPath)
		if err != nil {
			return fmt.Errorf("fs.Stat(%q): %w", evalPath, err)
		}

		// Skip other directories.
		if info.Mode().IsDir() && hostPath != root && hostPath != "/" {
			if err := walkRecursive(ctx, rootfs, tw, hostPath, creationTime, platform); err != nil {
				return err
			}
			continue
		}

		// Open the file to copy it into the tarball.
		log.V(4).Info("adding file to tar")
		file, err := rootfs.Open(evalPath)
		if err != nil {
			return fmt.Errorf("os.Open(%q): %w", evalPath, err)
		}

		// Copy the file into the image tarball.
		header := &tar.Header{
			Name:     hostPath,
			Size:     info.Size(),
			Typeflag: tar.TypeReg,
			Uid:      uid,
			Gid:      0,
			Mode:     int64(info.Mode()),
			ModTime:  creationTime.Time,
		}
		if err := tw.WriteHeader(header); err != nil {
			_ = file.Close()
			return fmt.Errorf("tar.Writer.WriteHeader(%q): %w", hostPath, err)
		}
		if _, err := io.Copy(tw, file); err != nil {
			_ = file.Close()
			return fmt.Errorf("io.Copy(%q, %q): %w", hostPath, evalPath, err)
		}
		_ = file.Close()
	}
	return nil
}
