package pipelines

import (
	"fmt"
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/go-logr/logr"
)

type Env struct{}

func (*Env) Run(ctx *BuildContext, options cbev1.Options) error {
	log := logr.FromContextOrDiscard(ctx.Context)

	for k, v := range options {
		log.V(5).Info("setting environment variable", "key", k, "value", v)
		ctx.Config.Env = append(ctx.Config.Env, fmt.Sprintf("%s=%v", k, v))
	}
	return nil
}

func (*Env) Name() string {
	return "env"
}
