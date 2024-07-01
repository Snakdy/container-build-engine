package pipelines

import (
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/go-logr/logr"
	"path/filepath"
)

type SymbolicLink struct {
	options cbev1.Options
}

func (s *SymbolicLink) Run(ctx *BuildContext, _ ...cbev1.Options) (cbev1.Options, error) {
	log := logr.FromContextOrDiscard(ctx.Context)

	for k, v := range s.options {
		srcPath := filepath.Clean(k)
		dstPath := filepath.Clean(v.(string))

		log.V(5).Info("creating link", "src", srcPath, "dst", dstPath)
		if err := ctx.FS.Symlink(srcPath, dstPath); err != nil {
			log.Error(err, "failed to create link")
			return cbev1.Options{}, err
		}
	}
	return cbev1.Options{}, nil
}

func (*SymbolicLink) Name() string {
	return StatementSymbolicLink
}

func (s *SymbolicLink) SetOptions(options cbev1.Options) {
	if s.options == nil {
		s.options = map[string]any{}
	}
	utils.CopyMap(options, s.options)
}
