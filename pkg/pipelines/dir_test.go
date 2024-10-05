package pipelines

import (
	"chainguard.dev/apko/pkg/apk/fs"
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

// interface guard
var _ PipelineStatement = &Dir{}

func TestDir_Run(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))
	wd, err := os.Getwd()
	require.NoError(t, err)

	var cases = []struct {
		name         string
		src          string
		dst          string
		expectedPath string
	}{
		{
			"local directory with shortform",
			"testdata",
			"/tmp/",
			"/tmp/testdata/text.txt",
		},
		{
			"local directory",
			"testdata",
			"/tmp/foo",
			"/tmp/foo/text.txt",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rootfs := fs.NewMemFS()
			s := &Dir{options: map[string]any{
				"src": tt.src,
				"dst": tt.dst,
			}}
			_, err := s.Run(&BuildContext{
				WorkingDirectory: wd,
				Context:          ctx,
				FS:               rootfs,
				ConfigFile: &v1.ConfigFile{
					Config: v1.Config{},
				},
			})
			assert.NoError(t, err)
			_, err = rootfs.Stat(tt.expectedPath)
			assert.NotErrorIs(t, err, os.ErrNotExist)
		})
	}
}
