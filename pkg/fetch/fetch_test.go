package fetch

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestChecksumFile(t *testing.T) {
	t.Run("without type prefix", func(t *testing.T) {
		err := checksumFile("./testdata/test.txt", "5067772cf39f7f42a7b5cd5d3b13da459fc9530b09722ae8fedc57dbbc0c50a3")
		assert.NoError(t, err)
	})
	t.Run("with type prefix", func(t *testing.T) {
		err := checksumFile("./testdata/test.txt", "sha256:5067772cf39f7f42a7b5cd5d3b13da459fc9530b09722ae8fedc57dbbc0c50a3")
		assert.NoError(t, err)
	})
}

func TestFetch(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))

	out := t.TempDir()
	path, err := Fetch(ctx, "file://./testdata/test.txt?archive=false", out, "sha256:5067772cf39f7f42a7b5cd5d3b13da459fc9530b09722ae8fedc57dbbc0c50a3")
	assert.NoError(t, err)
	t.Logf("path: %s", path)

	assert.FileExists(t, path)
}
