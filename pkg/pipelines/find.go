package pipelines

import cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"

type StatementFinder = func(name string, options cbev1.Options) PipelineStatement

func Find(name string, options cbev1.Options) PipelineStatement {
	switch name {
	case StatementEnv:
		return &Env{Options: options}
	case StatementFile:
		return &File{Options: options}
	case StatementSymbolicLink:
		return &SymbolicLink{Options: options}
	default:
		return nil
	}
}
