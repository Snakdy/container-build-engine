package builder

import (
	"context"
	"fmt"
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/containers"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/chainguard-dev/go-apk/pkg/fs"
	"github.com/go-logr/logr"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

func NewBuilder(pipeline cbev1.Pipeline, workingDir string) (*Builder, error) {
	statements := make([]pipelines.PipelineStatement, len(pipeline.Statements))
	for i := range pipeline.Statements {
		statement := pipelines.Find(pipeline.Statements[i].Name, pipeline.Statements[i].Options)
		if statement == nil {
			return nil, fmt.Errorf("could not find statement '%s'", pipeline.Statements[i].Name)
		}
		statements[i] = statement
	}
	return NewBuilderFromStatements(pipeline.Base, workingDir, statements), nil
}

func NewBuilderFromStatements(baseRef, workingDir string, statements []pipelines.PipelineStatement) *Builder {
	return &Builder{
		baseRef:    baseRef,
		workingDir: workingDir,
		statements: statements,
	}
}

func (b *Builder) Build(ctx context.Context, platform *v1.Platform) (v1.Image, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.Info("building image")

	baseImage, err := containers.Pull(ctx, b.baseRef)
	if err != nil {
		return nil, err
	}

	cfg, err := baseImage.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("extracing config: %w", err)
	}
	cfg = cfg.DeepCopy()

	buildContext := &pipelines.BuildContext{
		Context:          ctx,
		WorkingDirectory: b.workingDir,
		FS:               fs.NewMemFS(),
		ConfigFile:       cfg,
	}

	// run the filesystem mutations
	if err := b.applyFSMutations(buildContext); err != nil {
		return nil, err
	}

	layer, err := containers.NewLayer(ctx, buildContext.FS, platform)
	if err != nil {
		return nil, fmt.Errorf("creating layer: %w", err)
	}

	// append our layer
	layers := []mutate.Addendum{
		{
			MediaType: types.OCILayer,
			Layer:     layer,
			History: v1.History{
				Author:    "",
				CreatedBy: "container-build-engine",
				Created:   v1.Time{},
			},
		},
	}
	withData, err := mutate.Append(baseImage, layers...)
	if err != nil {
		return nil, fmt.Errorf("appending layer: %w", err)
	}

	_ = b.applyPlatform(cfg, platform)

	// run the config mutations
	if err := b.applyConfigMutations(buildContext); err != nil {
		return nil, err
	}
	img, err := mutate.ConfigFile(withData, cfg)
	if err != nil {
		return nil, fmt.Errorf("mutating config: %w", err)
	}
	return img, nil
}

func (b *Builder) applyPlatform(cfg *v1.ConfigFile, platform *v1.Platform) error {
	// copy platform metadata
	cfg.OS = platform.OS
	cfg.Architecture = platform.Architecture
	cfg.OSVersion = platform.OSVersion
	cfg.Variant = platform.Variant
	cfg.OSFeatures = platform.OSFeatures

	return nil
}

func (b *Builder) applyFSMutations(ctx *pipelines.BuildContext) error {
	log := logr.FromContextOrDiscard(ctx.Context)
	log.Info("applying fs mutation pipelines")
	for _, pipeline := range b.statements {
		if !pipeline.MutatesFS() {
			continue
		}
		if err := pipeline.Run(ctx); err != nil {
			return fmt.Errorf("running pipeline: %w", err)
		}
	}
	return nil
}

func (b *Builder) applyConfigMutations(ctx *pipelines.BuildContext) error {
	log := logr.FromContextOrDiscard(ctx.Context)
	log.Info("applying config mutation pipelines")
	for _, pipeline := range b.statements {
		if !pipeline.MutatesConfig() {
			continue
		}
		if err := pipeline.Run(ctx); err != nil {
			return fmt.Errorf("running pipeline: %w", err)
		}
	}
	return nil
}
