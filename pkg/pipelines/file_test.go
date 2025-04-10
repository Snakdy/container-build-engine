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
		subPath      string
		executable   bool
		checksum     string
		expectedPath string
	}{
		{
			"https file",
			"https://curl.se/ca/cacert.pem",
			"/tmp/cacerts.crt",
			"",
			false,
			"",
			"/tmp/cacerts.crt",
		},
		{
			"local file with extension, long form",
			"file://testdata/text.txt",
			"/tmp/text.txt",
			"",
			false,
			"",
			"/tmp/text.txt",
		},
		{
			"local file with extension",
			"testdata/text.txt",
			"/tmp/text.txt",
			"",
			false,
			"",
			"/tmp/text.txt",
		},
		{
			"local file without extension",
			"testdata/config-file",
			"/tmp/file.conf",
			"",
			false,
			"",
			"/tmp/file.conf",
		},
		{
			"local executable file",
			"testdata/shell-script",
			"/tmp/shell-script",
			"",
			true,
			"",
			"/tmp/shell-script",
		},
		{
			"local executable file with extension",
			"testdata/shell-script.sh",
			"/tmp/",
			"",
			true,
			"",
			"/tmp/shell-script.sh",
		},
		{
			"local executable file with no destination filename",
			"testdata/shell-script",
			"/tmp/",
			"",
			true,
			"",
			"/tmp/shell-script",
		},
		{
			"remote zip file",
			"https://github.com/hashicorp/go-getter/releases/download/v1.7.8/go-getter_1.7.8_linux_amd64.zip",
			"/tmp/go-getter",
			"go-getter",
			true,
			"",
			"/tmp/go-getter",
		},
		{
			"remote tar",
			"https://github.com/google/go-containerregistry/releases/download/v0.20.3/go-containerregistry_Linux_x86_64.tar.gz",
			"/tmp/crane",
			"crane",
			true,
			"",
			"/tmp/crane",
		},
		{
			"remote tar under dir",
			"https://get.helm.sh/helm-v3.17.2-linux-amd64.tar.gz",
			"/tmp/helm",
			"linux-amd64/helm",
			true,
			"",
			"/tmp/helm",
		},
		{
			"remote file with checksum",
			"https://ftp.gnu.org/gnu/hello/hello-2.12.tar.gz",
			"/tmp/hello",
			"",
			false,
			"cf04af86dc085268c5f4470fbae49b18afbc221b78096aab842d934a76bad0ab",
			"/tmp/hello",
		},
		{
			"remote file with checksum",
			"https://ftp.gnu.org/gnu/hello/hello-2.12.tar.gz?checksum=cf04af86dc085268c5f4470fbae49b18afbc221b78096aab842d934a76bad0ab&archive=false",
			"/tmp/",
			"",
			false,
			"",
			"/tmp/hello-2.12.tar.gz",
		},
		{
			"remote archive unpacked",
			"https://ftp.gnu.org/gnu/hello/hello-2.12.tar.gz?checksum=cf04af86dc085268c5f4470fbae49b18afbc221b78096aab842d934a76bad0ab",
			"/tmp/",
			"",
			false,
			"",
			"/tmp/hello-2.12/README",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rootfs := fs.NewMemFS()
			s := &File{options: map[string]any{
				"uri":        tt.uri,
				"path":       tt.path,
				"executable": tt.executable,
				"sub-path":   tt.subPath,
			}}
			_, err := s.Run(&BuildContext{
				WorkingDirectory: wd,
				Context:          ctx,
				FS:               rootfs,
				ConfigFile: &v1.ConfigFile{
					Config: v1.Config{},
				},
			})
			require.NoError(t, err)
			info, err := rootfs.Stat(tt.expectedPath)
			require.NotErrorIs(t, err, os.ErrNotExist)
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
