package builder

import (
	"context"
	"github.com/Snakdy/container-build-engine/pkg/pipelines"
	"github.com/Snakdy/container-build-engine/pkg/vfs"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestDataChaining(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))

	wd, err := os.Getwd()
	require.NoError(t, err)

	platform, err := v1.ParsePlatform("linux/amd64")
	require.NoError(t, err)

	builder, err := NewBuilder(ctx, "scratch", []pipelines.OrderedPipelineStatement{
		{
			ID:        "generate-fake-data",
			Options:   map[string]any{},
			Statement: &FakeSrc{},
			DependsOn: nil,
		},
		{
			ID:        "use-fake-data",
			Options:   map[string]any{},
			Statement: &FakeDst{},
			DependsOn: []string{"generate-fake-data"},
		},
	}, Options{WorkingDir: wd, FS: vfs.NewVFS(t.TempDir())})
	assert.NoError(t, err)
	assert.NotNil(t, builder)

	img, err := builder.Build(ctx, platform)
	assert.NoError(t, err)
	assert.NotNil(t, img)
}
