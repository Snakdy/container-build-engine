package fetch

import (
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
