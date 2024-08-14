package builder

import (
	"chainguard.dev/apko/pkg/apk/fs"
	"context"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

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
				"HOST": "ftp.gnu.org",
			},
			Statement: &pipelines.Env{},
			DependsOn: nil,
		},
		{
			ID: "download-file",
			Options: map[string]any{
				"uri":  "https://${HOST}/gnu/hello/hello-2.12.tar.gz?checksum=cf04af86dc085268c5f4470fbae49b18afbc221b78096aab842d934a76bad0ab&archive=false",
				"path": "/hello-2.12.tar.gz",
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
	}, Options{WorkingDir: wd, FS: fs.NewMemFS()})
	assert.NoError(t, err)
	assert.NotNil(t, builder)

	img, err := builder.Build(ctx, platform)
	assert.NoError(t, err)
	assert.NotNil(t, img)

	cfg, err := img.ConfigFile()
	assert.NoError(t, err)

	assert.Contains(t, cfg.Config.Env, "FOO=bar")
}
