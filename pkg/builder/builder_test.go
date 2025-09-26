package builder

import (
	"context"
	"os"
	"testing"

	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/Snakdy/container-build-engine/pkg/vfs"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Build(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))

	wd, err := os.Getwd()
	require.NoError(t, err)

	platform, err := v1.ParsePlatform("linux/amd64")
	require.NoError(t, err)

	builder, err := NewBuilder(ctx, "ghcr.io/djcass44/nib/srv:v1.5.0", nil, Options{
		WorkingDir: wd,
	})
	require.NoError(t, err)

	img, err := builder.Build(ctx, platform)
	assert.NoError(t, err)
	assert.NotNil(t, img)
}

func TestNewBuilderFromStatements(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))

	wd, err := os.Getwd()
	require.NoError(t, err)

	platform, err := v1.ParsePlatform("linux/amd64")
	require.NoError(t, err)

	builder, err := NewBuilder(ctx, "scratch", []pipelines.OrderedPipelineStatement{
		{
			ID: "apply-env",
			Options: map[string]any{
				"FOO":  "bar",
				"HOST": "raw.githubusercontent.com",
			},
			Statement: &pipelines.Env{},
			DependsOn: nil,
		},
		{
			ID: "download-file",
			Options: map[string]any{
				"uri":  "https://${HOST}/Snakdy/container-build-engine/refs/heads/main/README.md",
				"path": "/README.md",
			},
			Statement: &pipelines.File{},
			DependsOn: []string{"apply-env"},
		},
		{
			ID: "copy-file",
			Options: map[string]any{
				"uri":  "./testdata/test.txt",
				"path": "/test.txt",
			},
			Statement: &pipelines.File{},
			DependsOn: nil,
		},
	}, Options{WorkingDir: wd, FS: vfs.NewVFS(t.TempDir())})
	assert.NoError(t, err)
	assert.NotNil(t, builder)

	img, err := builder.Build(ctx, platform)
	assert.NoError(t, err)
	assert.NotNil(t, img)

	cfg, err := img.ConfigFile()
	assert.NoError(t, err)

	assert.Contains(t, cfg.Config.Env, "FOO=bar")
}
