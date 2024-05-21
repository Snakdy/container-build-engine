package pipelines

import (
	"context"
	"github.com/chainguard-dev/go-apk/pkg/fs"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type BuildContext struct {
	Context          context.Context
	WorkingDirectory string
	FS               fs.FullFS
	ConfigFile       *v1.ConfigFile
}

type PipelineStatement interface {
	Run(ctx *BuildContext) error
	Name() string
	MutatesConfig() bool
	MutatesFS() bool
}

const (
	StatementSymbolicLink = "link"
	StatementFile         = "file"
	StatementEnv          = "env"
)
