package pipelines

import (
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/go-logr/logr"
	"path/filepath"
)

type SymbolicLink struct{}

func (*SymbolicLink) Run(ctx *BuildContext, options cbev1.Options) error {
	log := logr.FromContextOrDiscard(ctx.Context)

	for k, v := range options {
		srcPath := filepath.Clean(k)
		dstPath := filepath.Clean(v.(string))

		log.V(5).Info("creating link", "src", srcPath, "dst", dstPath)
		if err := ctx.FS.Symlink(srcPath, dstPath); err != nil {
			log.Error(err, "failed to create link")
			return err
		}
	}
	return nil
}

func (*SymbolicLink) Name() string {
	return "link"
}
