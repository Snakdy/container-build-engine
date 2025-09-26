package pipelines

import (
	"os"
	"path/filepath"
	"strings"

	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/envs"
	"github.com/Snakdy/container-build-engine/pkg/fetch"
	"github.com/Snakdy/container-build-engine/pkg/files"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/go-logr/logr"
)

// File downloads or adds a file.
// Accepts the following parameters:
//
// 1. "path": where to place the file in the container
//
// 2. "uri": URI indicating where to get the file from. Supports https:// and file:// schemes and will
// default to file:// if none is provided.
//
// 3. "executable": make the file executable
//
// 4. "sub-path": if the file is an archive, extract a file from it
//
// 5. "checksum": hash of the file for checksum validation
type File struct {
	options cbev1.Options
}

func (s *File) Run(ctx *BuildContext, _ ...cbev1.Options) (cbev1.Options, error) {
	log := logr.FromContextOrDiscard(ctx.Context)
	log.V(7).Info("running statement", "options", s.options)

	rawPath, err := cbev1.GetRequired[string](s.options, "path")
	if err != nil {
		return cbev1.Options{}, err
	}
	fileUri, err := cbev1.GetRequired[string](s.options, "uri")
	if err != nil {
		return cbev1.Options{}, err
	}
	executable, err := cbev1.GetOptional[bool](s.options, "executable")
	if err != nil {
		return cbev1.Options{}, err
	}
	subPath, err := cbev1.GetOptional[string](s.options, "sub-path")
	if err != nil {
		return cbev1.Options{}, err
	}
	checksum, err := cbev1.GetOptional[string](s.options, "checksum")
	if err != nil {
		return cbev1.Options{}, err
	}

	// expand paths using environment variables
	path := filepath.Clean(envs.ExpandEnvFunc(rawPath, ExpandList(ctx.ConfigFile.Config.Env)))
	dst, err := os.MkdirTemp("", "file-*")
	if err != nil {
		log.Error(err, "failed to prepare download directory")
		return cbev1.Options{}, err
	}
	srcUri := envs.ExpandEnv(fileUri)

	log.V(2).Info("retrieving file", "file", srcUri, "path", dst)

	dst, err = fetch.Fetch(ctx.Context, srcUri, dst, checksum)
	if err != nil {
		log.Error(err, "failed to retrieve file", "src", srcUri, "dst", dst)
		return cbev1.Options{}, err
	}
	copySrc := dst

	dir := false
	if info, err := os.Stat(copySrc); err == nil {
		dir = info.IsDir()
		if dir {
			log.V(7).Info("source is a directory", "src", copySrc)
		}
	}

	if subPath != "" && dir {
		copySrc = filepath.Join(dst, subPath)
	}

	if executable {
		log.V(6).Info("setting executable bit", "file", copySrc)
		if err := os.Chmod(copySrc, 0755); err != nil {
			log.Error(err, "failed to update file permissions", "file", copySrc)
			return cbev1.Options{}, err
		}
	}

	dstDir := false
	if info, err := ctx.FS.Stat(path); err == nil {
		dstDir = info.IsDir()
		if dstDir {
			log.V(7).Info("destination is a directory", "dst", path)
		} else {
			log.V(7).Info("destination is a file", "dst", path)
		}
	} else {
		log.Error(err, "failed to stat file", "dst", path)
	}

	// handle short-form destination paths
	if (strings.HasSuffix(rawPath, "/") || dstDir) && !dir {
		path = filepath.Join(path, filepath.Base(copySrc))
	}

	log.V(5).Info("copying file or directory", "src", copySrc, "dst", path)
	if err := files.CopyDirectory(ctx.Context, copySrc, path, ctx.FS); err != nil {
		log.Error(err, "failed to copy directory", "src", copySrc, "dst", path)
		return cbev1.Options{}, err
	}
	return cbev1.Options{}, nil
}

func ExpandList(vs []string) func(s string) string {
	return func(s string) string {
		for _, e := range vs {
			k, v, _ := strings.Cut(e, "=")
			if k == s {
				return v
			}
		}
		return ""
	}
}

func (*File) Name() string {
	return StatementFile
}

func (s *File) SetOptions(options cbev1.Options) {
	if s.options == nil {
		s.options = map[string]any{}
	}
	utils.CopyMap(options, s.options)
}
