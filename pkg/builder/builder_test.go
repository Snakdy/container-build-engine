package builder

import (
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

	builder := NewBuilderFromStatements("scratch", wd, []pipelines.PipelineStatement{
		&pipelines.Env{Options: map[string]any{
			"FOO": "bar",
		}},
		&pipelines.File{Options: map[string]any{
			"uri":  "https://ftp.gnu.org/gnu/hello/hello-2.12.tar.gz?checksum=cf04af86dc085268c5f4470fbae49b18afbc221b78096aab842d934a76bad0ab&archive=false",
			"path": "/hello-2.12.tar.gz",
		}},
		&pipelines.File{Options: map[string]any{
			"uri":  "./testdata/test.txt",
			"path": "/test.txt",
		}},
	})

	img, err := builder.Build(ctx, platform)
	assert.NoError(t, err)
	assert.NotNil(t, img)

	cfg, err := img.ConfigFile()
	assert.NoError(t, err)

	assert.Contains(t, cfg.Config.Env, "FOO=bar")
}
