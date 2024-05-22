package builder

import (
	"context"
	"fmt"
	"github.com/Snakdy/container-build-engine/pkg/containers"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/stategraph"
	"github.com/Snakdy/container-build-engine/pkg/useradd"
	"github.com/chainguard-dev/go-apk/pkg/fs"
	"github.com/go-logr/logr"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"path/filepath"
	"slices"
	"strings"
)

const (
	DefaultPath     = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/home/somebody/.local/bin:/home/somebody/bin"
	DefaultUsername = "somebody"
)

func NewBuilder(ctx context.Context, baseRef, workingDir string, statements []pipelines.OrderedPipelineStatement) (*Builder, error) {
	log := logr.FromContextOrDiscard(ctx)

	// assemble the statement graph, so we know what
	// order to run them in
	graph := stategraph.New()
	for i := range statements {
		if err := graph.DependOn(statements[i]); err != nil {
			return nil, err
		}
	}
	orderedNames := graph.TopoSorted()

	// locate the statements and provide them with
	// their options
	orderedStatements := make([]pipelines.PipelineStatement, len(orderedNames))
	log.V(3).Info("assembled statement dependency graph", "orderedNames", orderedNames)
	for i := range orderedNames {
		// get the actual statement
		idx := slices.IndexFunc(statements, func(statement pipelines.OrderedPipelineStatement) bool {
			return statement.ID == orderedNames[i]
		})
		if idx < 0 {
			return nil, fmt.Errorf("could not locate statement in dependency tree")
		}
		orderedStatements[i] = statements[idx].Statement
		orderedStatements[i].SetOptions(statements[idx].Options)
	}
	return &Builder{
		baseRef:    baseRef,
		workingDir: workingDir,
		statements: orderedStatements,
	}, nil
}

func (b *Builder) Build(ctx context.Context, platform *v1.Platform) (v1.Image, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.Info("building image")

	baseImage, err := containers.Get(ctx, b.baseRef)
	if err != nil {
		return nil, err
	}

	if mt, err := baseImage.MediaType(); err == nil {
		log.V(3).Info("detected base image media type", "mediaType", mt)
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
	if err := b.applyMutations(buildContext); err != nil {
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
	// write the config file into the base image so that
	// appending works later on
	mutatedBase, err := mutate.ConfigFile(baseImage, buildContext.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("mutating config: %w", err)
	}

	// append our layer
	log.V(3).Info("appending layer")
	withData, err := mutate.Append(mutatedBase, mutate.Addendum{
		MediaType: types.OCILayer,
		Layer:     layer,
		History: v1.History{
			Author:    "",
			CreatedBy: "container-build-engine",
			Created:   v1.Time{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("appending layer: %w", err)
	}
	cfg, err = withData.ConfigFile()
	if err != nil {
		return nil, err
	}

	b.applyPlatform(cfg, platform)
	b.applyPath(cfg)

	// package everything up
	img, err := mutate.ConfigFile(withData, cfg)
	if err != nil {
		return nil, fmt.Errorf("mutating config: %w", err)
	}
	return img, nil
}

// applyPath sets the PATH environment variable.
// If the variable already exists, it appends to it
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

func (b *Builder) applyMutations(ctx *pipelines.BuildContext) error {
	log := logr.FromContextOrDiscard(ctx.Context)
	log.Info("applying mutation pipelines")
	for _, pipeline := range b.statements {
		if err := pipeline.Run(ctx); err != nil {
			return fmt.Errorf("running pipeline: %w", err)
		}
	}
	return nil
}
