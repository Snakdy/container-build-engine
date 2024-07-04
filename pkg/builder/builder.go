package builder

import (
	"context"
	"fmt"
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/containers"
	"github.com/Snakdy/container-build-engine/pkg/envs"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/stategraph"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/Snakdy/container-build-engine/pkg/useradd"
	"github.com/chainguard-dev/go-apk/pkg/fs"
	"github.com/go-logr/logr"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const DefaultUsername = "somebody"
const DefaultUid = 1001

func NewBuilder(ctx context.Context, baseRef string, statements []pipelines.OrderedPipelineStatement, options Options) (*Builder, error) {
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
		options:    options,
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

	var filesystem fs.FullFS

	if b.options.DirFS {
		tmpFs, err := os.MkdirTemp("", "container-build-engine-fs-*")
		if err != nil {
			return nil, fmt.Errorf("creating temporary directory: %w", err)
		}
		log.V(3).Info("creating tempfs virtual filesystem", "path", tmpFs)
		filesystem = fs.DirFS(tmpFs)
	} else {
		filesystem = fs.NewMemFS()
		log.V(3).Info("creating in-memory virtual filesystem - this may cause memory issues with large builds")
	}

	buildContext := &pipelines.BuildContext{
		Context:          ctx,
		WorkingDirectory: b.options.WorkingDir,
		FS:               filesystem,
		ConfigFile:       cfg,
	}

	// create the non-root user directory
	buildContext.ConfigFile.Config.Env = pipelines.SetOrAppend(buildContext.ConfigFile.Config.Env, "HOME", filepath.Join("/home", b.options.GetUsername()))
	if err := useradd.NewUserDir(ctx, buildContext.FS, b.options.GetUsername(), b.options.GetUid()); err != nil {
		return nil, err
	}

	// run the filesystem mutations
	if err := b.applyMutations(buildContext); err != nil {
		return nil, err
	}

	// create the non-root user
	if err := useradd.NewUser(ctx, buildContext.FS, b.options.GetUsername(), b.options.GetUid()); err != nil {
		return nil, err
	}

	layer, err := containers.NewLayer(ctx, buildContext.FS, b.options.GetUsername(), b.options.GetUid(), platform)
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
			Author:    b.options.Metadata.Author,
			CreatedBy: b.options.Metadata.GetCreatedBy(),
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

	b.applyPlatform(ctx, cfg, platform)
	b.applyPath(cfg)

	// package everything up
	img, err := mutate.ConfigFile(withData, cfg)
	if err != nil {
		return nil, fmt.Errorf("mutating config: %w", err)
	}
	// remove any randomness in the build
	// so that we can reproduce it
	canonicalImage, err := mutate.Canonical(img)
	if err != nil {
		return nil, fmt.Errorf("generating canonical image: %w", err)
	}
	return canonicalImage, nil
}

// applyPath sets the PATH environment variable.
// If the variable already exists, it appends to it
func (b *Builder) applyPath(cfg *v1.ConfigFile) {
	var found bool
	for i, e := range cfg.Config.Env {
		if strings.HasPrefix(e, "PATH=") {
			cfg.Config.Env[i] = cfg.Config.Env[i] + fmt.Sprintf(":%s:%s", filepath.Join("/home", b.options.GetUsername(), ".local", "bin"), filepath.Join("/home", b.options.GetUsername(), "bin"))
			found = true
		}
	}
	if !found {
		cfg.Config.Env = append(cfg.Config.Env, "PATH="+defaultPath(b.options.GetUsername()))
	}
}

func (b *Builder) applyPlatform(ctx context.Context, cfg *v1.ConfigFile, platform *v1.Platform) {
	log := logr.FromContextOrDiscard(ctx)

	// copy platform metadata
	cfg.OS = platform.OS
	cfg.Architecture = platform.Architecture
	cfg.OSVersion = platform.OSVersion
	cfg.Variant = platform.Variant
	cfg.OSFeatures = platform.OSFeatures

	// set the user
	cfg.Config.WorkingDir = filepath.Join("/home", b.options.GetUsername())
	cfg.Config.User = b.options.GetUsername()

	if cfg.Config.Labels == nil {
		cfg.Config.Labels = map[string]string{}
	}
	// todo support base image annotation hints
	// https://github.com/google/go-containerregistry/blob/main/cmd/crane/rebase.md#base-image-annotation-hints

	if b.options.Entrypoint != nil || b.options.ForceEntrypoint {
		for i := range b.options.Entrypoint {
			b.options.Entrypoint[i] = envs.ExpandEnvFunc(b.options.Entrypoint[i], pipelines.ExpandList(cfg.Config.Env))
		}
		log.V(4).Info("overriding entrypoint", "before", cfg.Config.Entrypoint, "after", b.options.Entrypoint)
		cfg.Config.Entrypoint = b.options.Entrypoint
	}
	if b.options.Command != nil || b.options.ForceEntrypoint {
		for i := range b.options.Command {
			b.options.Command[i] = envs.ExpandEnvFunc(b.options.Command[i], pipelines.ExpandList(cfg.Config.Env))
		}
		log.V(4).Info("overriding command", "before", cfg.Config.Cmd, "after", b.options.Command)
		cfg.Config.Cmd = b.options.Command
	}
}

func (b *Builder) applyMutations(ctx *pipelines.BuildContext) error {
	log := logr.FromContextOrDiscard(ctx.Context)
	log.Info("applying mutation pipelines")
	data := cbev1.Options{}
	for _, pipeline := range b.statements {
		out, err := pipeline.Run(ctx, data)
		if err != nil {
			return fmt.Errorf("running pipeline: %w", err)
		}
		utils.CopyMap(out, data)
	}
	return nil
}

func defaultPath(username string) string {
	return fmt.Sprintf("/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/home/%s/.local/bin:/home/%s/bin", username, username)
}
