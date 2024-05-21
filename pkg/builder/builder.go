package builder

import (
	"context"
	"fmt"
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/containers"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/Snakdy/container-build-engine/pkg/useradd"
	"github.com/chainguard-dev/go-apk/pkg/fs"
	"github.com/go-logr/logr"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"path/filepath"
	"strings"
)

const (
	DefaultPath     = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/home/somebody/.local/bin:/home/somebody/bin"
	DefaultUsername = "somebody"
)

func NewBuilder(pipeline cbev1.Pipeline, statementFinder pipelines.StatementFinder, workingDir string) (*Builder, error) {
	// if the user didn't specify a statement finder, we
	// need to use the default one
	if statementFinder == nil {
		statementFinder = pipelines.Find
	}
	// locate the statements and provide them with
	// their options
	statements := make([]pipelines.PipelineStatement, len(pipeline.Statements))
	for i := range pipeline.Statements {
		statement := statementFinder(pipeline.Statements[i].Name, pipeline.Statements[i].Options)
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

	baseImage, err := containers.Get(ctx, b.baseRef)
	if err != nil {
		return nil, err
	}

	cfg, err := baseImage.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("extracting config: %w", err)
	}
	cfg = cfg.DeepCopy()

	buildContext := &pipelines.BuildContext{
		Context:          ctx,
		WorkingDirectory: b.workingDir,
		FS:               fs.NewMemFS(),
		ConfigFile:       cfg,
	}

	// create the non-root user
	if err := useradd.NewUser(ctx, buildContext.FS, DefaultUsername, 1001); err != nil {
		return nil, err
	}

	// run the filesystem mutations
	if err := b.applyFSMutations(buildContext); err != nil {
		return nil, err
	}

	layer, err := containers.NewLayer(ctx, buildContext.FS, DefaultUsername, platform)
	if err != nil {
		return nil, fmt.Errorf("creating layer: %w", err)
	}

	// convert the base image to OCI format
	if mt, err := baseImage.MediaType(); err == nil {
		log.V(1).Info("detected base image media type", "mediaType", mt)
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

	b.applyPlatform(cfg, platform)

	// run the config mutations
	if err := b.applyConfigMutations(buildContext); err != nil {
		return nil, err
	}

	// package everything up
	img, err := mutate.ConfigFile(withData, cfg)
	if err != nil {
		return nil, fmt.Errorf("mutating config: %w", err)
	}
	return img, nil
}

func (b *Builder) applyPath(cfg *v1.ConfigFile) {
	var found bool
	for i, e := range cfg.Config.Env {
		if strings.HasPrefix(e, "PATH=") {
			cfg.Config.Env[i] = cfg.Config.Env[i] + fmt.Sprintf(":%s:%s", filepath.Join("/home", DefaultUsername, ".local", "bin"), filepath.Join("/home", DefaultUsername, "bin"))
			found = true
		}
	}
	if !found {
		cfg.Config.Env = append(cfg.Config.Env, "PATH="+DefaultPath)
	}
}

func (b *Builder) applyPlatform(cfg *v1.ConfigFile, platform *v1.Platform) {
	// copy platform metadata
	cfg.OS = platform.OS
	cfg.Architecture = platform.Architecture
	cfg.OSVersion = platform.OSVersion
	cfg.Variant = platform.Variant
	cfg.OSFeatures = platform.OSFeatures

	// set the user
	cfg.Config.WorkingDir = filepath.Join("/home", DefaultUsername)
	cfg.Config.User = DefaultUsername

	if cfg.Config.Labels == nil {
		cfg.Config.Labels = map[string]string{}
	}
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
