package containers

import (
	"context"
	"github.com/Snakdy/container-build-engine/pkg/oci/empty"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))

	t.Run("real image", func(t *testing.T) {
		img, err := Get(ctx, "busybox")
		assert.NoError(t, err)
		size, err := img.Size()
		assert.NoError(t, err)
		assert.NotZero(t, size)
	})
	t.Run("scratch image", func(t *testing.T) {
		img, err := Get(ctx, "scratch")
		assert.NoError(t, err)
		size, err := img.Size()
		assert.NoError(t, err)
		assert.NotZero(t, size)

		assert.Equal(t, empty.Image, img)
	})
}
