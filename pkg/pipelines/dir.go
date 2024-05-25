package pipelines

import (
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/envs"
	"github.com/Snakdy/container-build-engine/pkg/files"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/go-logr/logr"
	"path/filepath"
)

type Dir struct {
	options cbev1.Options
}

func (s *Dir) Run(ctx *BuildContext) error {
	log := logr.FromContextOrDiscard(ctx.Context)
	log.V(7).Info("running statement", "options", s.options)

	src, err := cbev1.GetRequired[string](s.options, "src")
	if err != nil {
		return err
	}
	dst, err := cbev1.GetRequired[string](s.options, "dst")
	if err != nil {
		return err
	}

	// expand paths
	src = filepath.Clean(envs.ExpandEnvFunc(src, ExpandList(ctx.ConfigFile.Config.Env)))
	dst = filepath.Clean(envs.ExpandEnvFunc(dst, ExpandList(ctx.ConfigFile.Config.Env)))

	// copy the directory
	if err := files.CopyDirectory(src, dst, ctx.FS); err != nil {
		log.Error(err, "failed to copy directory")
		return err
	}

	return nil
}

func (*Dir) Name() string {
	return StatementDir
}

func (s *Dir) SetOptions(options cbev1.Options) {
	if s.options == nil {
		s.options = map[string]any{}
	}
	utils.CopyMap(options, s.options)
}
