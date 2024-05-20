package containers

import (
	"context"
	"github.com/Snakdy/container-build-engine/pkg/oci/empty"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestSave(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))

	out := filepath.Join(t.TempDir(), "test.tar")

	err := Save(ctx, empty.Image, "scratch:latest", out)
	assert.NoError(t, err)

	assert.FileExists(t, out)
}
