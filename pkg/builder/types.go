package builder

import (
	"chainguard.dev/apko/pkg/apk/fs"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
)

type Builder struct {
	baseRef    string
	options    Options
	statements []pipelines.PipelineStatement
}

type Options struct {
	WorkingDir string
	// Username is the name of the Linux user
	// that the container will run as.
	Username string
	// Uid is the numerical ID of the Linux user
	// that the container will run as.
	Uid int
	// Shell is the default shell that will be opened
	// whenever a user connects. If not provided
	// it will default to /bin/sh
	Shell           string
	Entrypoint      []string
	Command         []string
	ForceEntrypoint bool
	Metadata        MetadataOptions
	FS              fs.FullFS
}

type MetadataOptions struct {
	Author    string
	CreatedBy string
}

// GetUsername returns the nominated username or the
// DefaultUsername
func (o *Options) GetUsername() string {
	if o.Username == "" {
		return DefaultUsername
	}
	return o.Username
}

// GetUid returns the nominated uid or the DefaultUid
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
