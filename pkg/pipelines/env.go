package pipelines

import (
	"fmt"
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/go-logr/logr"
)

type Env struct {
	Options cbev1.Options
}

func (s *Env) Run(ctx *BuildContext) error {
	log := logr.FromContextOrDiscard(ctx.Context)

	for k, v := range s.Options {
		log.V(5).Info("setting environment variable", "key", k, "value", v)
		ctx.ConfigFile.Config.Env = append(ctx.ConfigFile.Config.Env, fmt.Sprintf("%s=%v", k, v))
	}
	return nil
}

func (*Env) Name() string {
	return StatementEnv
}

func (*Env) MutatesConfig() bool {
	return true
}

func (*Env) MutatesFS() bool {
	return false
}
