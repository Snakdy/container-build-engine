package pipelines

import cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"

type StatementFinder = func(name string, options cbev1.Options) PipelineStatement

func Find(name string, options cbev1.Options) PipelineStatement {
	var s PipelineStatement
	switch name {
	case StatementEnv:
		s = &Env{}
	case StatementFile:
		s = &File{}
	case StatementSymbolicLink:
		s = &SymbolicLink{}
	default:
		return nil
	}
	s.SetOptions(options)
	return s
}
