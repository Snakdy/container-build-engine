package builder

import "github.com/Snakdy/container-build-engine/pkg/pipelines"

type Builder struct {
	baseRef    string
	statements []pipelines.PipelineStatement
}
