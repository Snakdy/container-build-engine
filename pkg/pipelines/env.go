package pipelines

import (
	"fmt"
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/envs"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/go-logr/logr"
	"os"
	"slices"
	"strings"
)

// Env exports one or more environment variables. Options should be a key-value map
// where key is the name and the value is the value to set.
type Env struct {
	options cbev1.Options
}

func (s *Env) Run(ctx *BuildContext, _ ...cbev1.Options) (cbev1.Options, error) {
	log := logr.FromContextOrDiscard(ctx.Context)

	for k, v := range s.options {
		value := envs.ExpandEnvFunc(v.(string), ExpandList(ctx.ConfigFile.Config.Env))
		log.V(5).Info("exporting environment variable", "key", k, "value", v, "expandedValue", value)
		ctx.ConfigFile.Config.Env = SetOrAppend(ctx.ConfigFile.Config.Env, k, value)
		if err := os.Setenv(k, v.(string)); err != nil {
			log.Error(err, "could not export environment variable for usage in later stages", "key", k, "value", v, "expandedValue", value)
			return cbev1.Options{}, err
		}
	}
	return cbev1.Options{}, nil
}

func SetOrAppend(vars []string, k, v string) []string {
	idx := slices.IndexFunc(vars, func(s string) bool {
		return strings.HasPrefix(s, k+"=")
	})
	if idx < 0 {
		return append(vars, fmt.Sprintf("%s=%v", k, v))
	}
	vars[idx] = fmt.Sprintf("%s=%s", k, v)
	return vars
}

func (*Env) Name() string {
	return StatementEnv
}

func (s *Env) SetOptions(options cbev1.Options) {
	if s.options == nil {
		s.options = map[string]any{}
	}
	utils.CopyMap(options, s.options)
}
