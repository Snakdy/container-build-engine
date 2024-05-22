package pipelines

import (
	"fmt"
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/go-logr/logr"
	"os"
)

type Env struct {
	options cbev1.Options
}

func (s *Env) Run(ctx *BuildContext) error {
	log := logr.FromContextOrDiscard(ctx.Context)

	for k, v := range s.options {
		log.V(5).Info("exporting environment variable", "key", k, "value", v)
		ctx.ConfigFile.Config.Env = append(ctx.ConfigFile.Config.Env, fmt.Sprintf("%s=%v", k, v))
		if err := os.Setenv(k, v.(string)); err != nil {
			log.Error(err, "could not export environment variable for usage in later stages", "key", k, "value", v)
			return err
		}
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

func (s *Env) SetOptions(options cbev1.Options) {
	if s.options == nil {
		s.options = map[string]any{}
	}
	utils.CopyMap(options, s.options)
}
