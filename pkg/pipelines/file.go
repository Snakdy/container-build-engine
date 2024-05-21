package pipelines

import (
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/envs"
	"github.com/Snakdy/container-build-engine/pkg/files"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-getter"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	Options cbev1.Options
}

func (s *File) Run(ctx *BuildContext) error {
	log := logr.FromContextOrDiscard(ctx.Context)

	path, err := s.Options.GetString("path")
	if err != nil {
		return err
	}
	fileUri, err := s.Options.GetString("uri")
	if err != nil {
		return err
	}
	executable, err := s.Options.GetOptionalBool("executable")
	if err != nil {
		return err
	}
	subPath, err := s.Options.GetOptionalString("sub-path")
	if err != nil {
		return err
	}

	// expand paths using environment variables
	path = filepath.Clean(os.Expand(path, expandList(ctx.ConfigFile.Config.Env)))
	dst, err := os.MkdirTemp("", "file-*")
	if err != nil {
		log.Error(err, "failed to prepare download directory")
		return err
	}
	srcUri, err := url.Parse(envs.ExpandEnv(fileUri))
	if err != nil {
		return err
	}
	//q := srcUri.Query()
	//q.Set("checksum", checksum.Integrity)
	//srcUri.RawQuery = q.Encode()

	log.Info("retrieving file", "file", srcUri.String(), "path", dst)
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
		log.Error(err, "failed to retrieve file", "src", srcUri.String())
		return err
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
			return err
		}
	}
	// todo update file permissions for file types that don't match the above
	log.V(5).Info("copying file or directory", "src", copySrc, "dst", path)
	if err := files.CopyDirectory(copySrc, path, ctx.FS); err != nil {
		log.Error(err, "failed to copy directory")
		return err
	}
	return nil
}

func expandList(vs []string) func(s string) string {
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

func (*File) MutatesConfig() bool {
	return false
}

func (*File) MutatesFS() bool {
	return true
}
