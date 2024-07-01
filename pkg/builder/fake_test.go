package builder

import (
	"fmt"
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/go-logr/logr"
)

type FakeSrc struct {
	options cbev1.Options
}

func (s *FakeSrc) Run(*pipelines.BuildContext, ...cbev1.Options) (cbev1.Options, error) {
	return cbev1.Options{
		"src": "test",
	}, nil
}

func (*FakeSrc) Name() string {
	return "src"
}

func (s *FakeSrc) SetOptions(options cbev1.Options) {
	if s.options == nil {
		s.options = map[string]any{}
	}
	utils.CopyMap(options, s.options)
}

type FakeDst struct {
	options cbev1.Options
}

func (s *FakeDst) Run(ctx *pipelines.BuildContext, runtimeOptions ...cbev1.Options) (cbev1.Options, error) {
	log := logr.FromContextOrDiscard(ctx.Context)
	val, err := cbev1.GetAny[string](runtimeOptions, "src")
	if err != nil {
		return nil, err
	}
	log.Info("detected data", "value", val)
	if val != "test" {
		return nil, fmt.Errorf("invalid value %s", val)
	}
	return cbev1.Options{}, nil
}

func (*FakeDst) Name() string {
	return "dst"
}

func (s *FakeDst) SetOptions(options cbev1.Options) {
	if s.options == nil {
		s.options = map[string]any{}
	}
	utils.CopyMap(options, s.options)
}
