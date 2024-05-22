package builder

import "github.com/Snakdy/container-build-engine/pkg/pipelines"

type Builder struct {
	baseRef    string
	options    Options
	statements []pipelines.PipelineStatement
}

type Options struct {
	WorkingDir string
	Username   string
}

func (o *Options) GetUsername() string {
	if o.Username == "" {
		return DefaultUsername
	}
	return o.Username
}
