package envs

import (
	"github.com/drone/envsubst"
)

func ExpandEnv(s string) string {
	val, err := envsubst.EvalEnv(s)
	if err != nil {
		return s
	}
	return val
}

func ExpandEnvFunc(s string, mapping func(s string) string) string {
	val, err := envsubst.Eval(s, mapping)
	if err != nil {
		return s
	}
	return val
}
