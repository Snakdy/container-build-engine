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
	ConfigFile       *v1.ConfigFile
}

type PipelineStatement interface {
	Run(ctx *BuildContext) error
	Name() string
	SetOptions(options cbev1.Options)
}

type OrderedPipelineStatement struct {
	ID        string
	Options   cbev1.Options
	Statement PipelineStatement
	DependsOn []string
}

const (
	StatementSymbolicLink = "link"
	StatementFile         = "file"
	StatementEnv          = "env"
	StatementScript       = "script"
	StatementDir          = "dir"
)
