package pipelines

import (
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/envs"
	"github.com/Snakdy/container-build-engine/pkg/files"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/go-logr/logr"
	"path/filepath"
)

// Dir recursively copies a directory into the container.
// Accepts the following parameters:
//
// 1. "src": where to retrieve the directory from
//
// 2. "dst": where to place the directory in the container
type Dir struct {
	options cbev1.Options
}

func (s *Dir) Run(ctx *BuildContext, runtimeOptions ...cbev1.Options) (cbev1.Options, error) {
	log := logr.FromContextOrDiscard(ctx.Context)
	log.V(7).Info("running statement", "options", s.options)

	options := append(cbev1.OptionsList{s.options}, runtimeOptions...)

	src, err := cbev1.GetAny[string](options, "src")
	if err != nil {
		return cbev1.Options{}, err
	}
	dst, err := cbev1.GetAny[string](options, "dst")
	if err != nil {
		return cbev1.Options{}, err
	}

	// expand paths
	src = filepath.Clean(envs.ExpandEnvFunc(src, ExpandList(ctx.ConfigFile.Config.Env)))
	dst = filepath.Clean(envs.ExpandEnvFunc(dst, ExpandList(ctx.ConfigFile.Config.Env)))

	// copy the directory
	if err := files.CopyDirectory(src, dst, ctx.FS); err != nil {
		log.Error(err, "failed to copy directory")
		return cbev1.Options{}, err
	}

	return cbev1.Options{}, nil
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
