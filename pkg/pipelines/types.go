package pipelines

import (
	"context"
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/chainguard-dev/go-apk/pkg/fs"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type BuildContext struct {
	Context          context.Context
	WorkingDirectory string
	FS               fs.FullFS
	Config           *v1.Config
}

type PipelineStatement interface {
	Run(ctx *BuildContext, options cbev1.Options) error
	Name() string
}
