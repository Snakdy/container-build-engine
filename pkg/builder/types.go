package builder

import (
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
)

type Builder struct {
	baseRef    string
	options    Options
	statements []pipelines.PipelineStatement
}

type Options struct {
	WorkingDir      string
	Username        string
	Uid             int
	Entrypoint      []string
	Command         []string
	ForceEntrypoint bool
	Metadata        MetadataOptions
	DirFS           bool
}

type MetadataOptions struct {
	Author    string
	CreatedBy string
}

func (o *Options) GetUsername() string {
	if o.Username == "" {
		return DefaultUsername
	}
	return o.Username
}

func (o *Options) GetUid() int {
	if o.Uid <= 0 {
		return DefaultUid
	}
	return o.Uid
}

func (o *MetadataOptions) GetCreatedBy() string {
	if o.CreatedBy == "" {
		return "container-build-engine"
	}
	return o.CreatedBy
}
