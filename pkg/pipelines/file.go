package pipelines

import (
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/envs"
	"github.com/Snakdy/container-build-engine/pkg/files"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-getter"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// File downloads or adds a file.
// Accepts the following parameters:
//
// 1. "path": where to place the file in the container
//
// 2. "uri": go-getter style URI indicating where to get the file from
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
	srcUri, err := url.Parse(envs.ExpandEnv(fileUri))
	if err != nil {
		return cbev1.Options{}, err
	}
	if checksum != "" {
		log.V(2).Info("validating checksum", "uri", fileUri, "checksum", checksum)
		q := srcUri.Query()
		q.Set("checksum", checksum)
		srcUri.RawQuery = q.Encode()
	}

	log.V(2).Info("retrieving file", "file", srcUri.String(), "path", dst)
	client := &getter.Client{
		Ctx:             ctx.Context,
		Pwd:             ctx.WorkingDirectory,
		Src:             srcUri.String(),
		Dst:             dst,
		DisableSymlinks: true,
		Mode:            getter.ClientModeAny,
		Getters:         getters,
	}
	if err := client.Get(); err != nil {
		log.Error(err, "failed to retrieve file", "src", srcUri.String(), "dst", dst)
		return cbev1.Options{}, err
	}
	var permissions os.FileMode = 0644
	if executable {
		permissions = 0755
	}
	copySrc := dst
	if subPath != "" || filepath.Ext(fileUri) == "" {
		if subPath != "" {
			copySrc = filepath.Join(dst, subPath)
		}
		if filepath.Ext(fileUri) == "" {
			copySrc = filepath.Join(dst, filepath.Base(fileUri))
		}
		log.V(6).Info("updating file permissions", "file", copySrc, "permissions", permissions)
		if err := os.Chmod(copySrc, permissions); err != nil {
			log.Error(err, "failed to update file permissions", "file", copySrc)
			return cbev1.Options{}, err
		}
	}

	// handle short-form destination paths
	if strings.HasSuffix(rawPath, "/") {
		path = filepath.Join(path, filepath.Base(copySrc))
	}

	// todo update file permissions for file types that don't match the above
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

var getters = map[string]getter.Getter{
	"file":  &getter.FileGetter{Copy: true},
	"https": &getter.HttpGetter{XTerraformGetDisabled: true, Netrc: true},
	"s3":    &getter.S3Getter{},
	"git":   &getter.GitGetter{},
	"gcs":   &getter.GCSGetter{},
	"hg":    &getter.HgGetter{},
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
