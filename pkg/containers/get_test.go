package containers

import (
	"context"
	"github.com/Snakdy/container-build-engine/pkg/oci/empty"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))

	require.NoError(t, os.Setenv("TMPDIR", t.TempDir()))

	t.Run("real image", func(t *testing.T) {
		init := time.Now()
		img, err := Get(ctx, "alpine/xml")
		assert.NoError(t, err)
		control := time.Since(init)
		size, err := img.Size()
		assert.NoError(t, err)
		assert.NotZero(t, size)

		t.Logf("control: %s", control)

		for i := 0; i < 5; i++ {
			start := time.Now()
			img, err = Get(ctx, "alpine/xml")
			assert.NoError(t, err)
			timeTaken := time.Since(start)
			t.Logf("run %d: %s", i, timeTaken)

			assert.Less(t, timeTaken, control)
		}
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
