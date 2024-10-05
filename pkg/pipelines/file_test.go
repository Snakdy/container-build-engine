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
var _ PipelineStatement = &File{}

func TestFile_Run(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))
	wd, err := os.Getwd()
	require.NoError(t, err)

	var cases = []struct {
		name         string
		uri          string
		path         string
		executable   bool
		expectedPath string
	}{
		{
			"https file",
			"https://curl.se/ca/cacert.pem",
			"/tmp/cacerts.crt",
			false,
			"/tmp/cacerts.crt",
		},
		{
			"local file with extension",
			"testdata/text.txt",
			"/tmp/text.txt",
			false,
			"/tmp/text.txt",
		},
		{
			"local file without extension",
			"testdata/config-file",
			"/tmp/file.conf",
			false,
			"/tmp/file.conf",
		},
		{
			"local executable file",
			"testdata/shell-script",
			"/tmp/shell-script",
			true,
			"/tmp/shell-script",
		},
		{
			"local executable file with no destination filename",
			"testdata/shell-script",
			"/tmp/",
			true,
			"/tmp/shell-script",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rootfs := fs.NewMemFS()
			s := &File{options: map[string]any{
				"uri":        tt.uri,
				"path":       tt.path,
				"executable": tt.executable,
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
			info, err := rootfs.Stat(tt.expectedPath)
			assert.NotErrorIs(t, err, os.ErrNotExist)
			// https://stackoverflow.com/a/60128480
			if tt.executable {
				// executable by owner
				assert.True(t, info.Mode()&0100 != 0)
				// executable by group
				assert.True(t, info.Mode()&0010 != 0)
			}
		})
	}
}
