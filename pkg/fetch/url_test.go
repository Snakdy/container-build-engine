package fetch

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAmbientCredentials(t *testing.T) {
	// Backup and restore env
	origGithub := os.Getenv("GITHUB_TOKEN")
	origGitlab := os.Getenv("GITLAB_TOKEN")
	origJob := os.Getenv("CI_JOB_TOKEN")
	origServer := os.Getenv("CI_SERVER_URL")
	defer func() {
		_ = os.Setenv("GITHUB_TOKEN", origGithub)
		_ = os.Setenv("GITLAB_TOKEN", origGitlab)
		_ = os.Setenv("CI_JOB_TOKEN", origJob)
		_ = os.Setenv("CI_SERVER_URL", origServer)
	}()

	tests := []struct {
		name     string
		uri      string
		env      map[string]string
		expected http.Header
	}{
		{
			name: "GitHub Token",
			uri:  "https://github.com/foo/bar",
			env: map[string]string{
				"GITHUB_TOKEN": "gh-token",
			},
			expected: http.Header{
				"Authorization": []string{"Bearer gh-token"},
			},
		},
		{
			name: "GitLab Token (gitlab.com)",
			uri:  "https://gitlab.com/foo/bar",
			env: map[string]string{
				"GITLAB_TOKEN": "gl-token",
			},
			expected: http.Header{
				"Authorization": []string{"Bearer gl-token"},
				"Private-Token": []string{"gl-token"},
			},
		},
		{
			name: "GitLab Job Token",
			uri:  "https://my-gitlab.com/foo/bar",
			env: map[string]string{
				"CI_JOB_TOKEN":  "job-token",
				"CI_SERVER_URL": "https://my-gitlab.com",
			},
			expected: http.Header{
				"Authorization": []string{"Bearer job-token"},
				"Job-Token":     []string{"job-token"},
			},
		},
		{
			name: "GitLab Token (self-managed, detected via CI_SERVER_URL)",
			uri:  "https://my-gitlab.com/foo/bar",
			env: map[string]string{
				"GITLAB_TOKEN":  "gl-token",
				"CI_SERVER_URL": "https://my-gitlab.com",
			},
			expected: http.Header{
				"Authorization": []string{"Bearer gl-token"},
				"Private-Token": []string{"gl-token"},
			},
		},
		{
			name: "GitLab Token (self-managed, detected via legacy prefix)",
			uri:  "https://gitlab.my-company.com/foo/bar",
			env: map[string]string{
				"GITLAB_TOKEN": "gl-token",
			},
			expected: http.Header{
				"Authorization": []string{"Bearer gl-token"},
				"Private-Token": []string{"gl-token"},
			},
		},
		{
			name: "GitHub Token (via GH_TOKEN)",
			uri:  "https://github.com/foo/bar",
			env: map[string]string{
				"GH_TOKEN": "gh-token",
			},
			expected: http.Header{
				"Authorization": []string{"Bearer gh-token"},
			},
		},
		{
			name: "GitLab Token (via GL_TOKEN)",
			uri:  "https://gitlab.com/foo/bar",
			env: map[string]string{
				"GL_TOKEN": "gl-token",
			},
			expected: http.Header{
				"Authorization": []string{"Bearer gl-token"},
				"Private-Token": []string{"gl-token"},
			},
		},
		{
			name: "GitLab Precedence (GITLAB_TOKEN > GL_TOKEN)",
			uri:  "https://gitlab.com/foo/bar",
			env: map[string]string{
				"GITLAB_TOKEN": "gl-token-primary",
				"GL_TOKEN":     "gl-token-secondary",
			},
			expected: http.Header{
				"Authorization": []string{"Bearer gl-token-primary"},
				"Private-Token": []string{"gl-token-primary"},
			},
		},
		{
			name: "GitLab Precedence (GITLAB_TOKEN > CI_JOB_TOKEN)",
			uri:  "https://gitlab.com/foo/bar",
			env: map[string]string{
				"GITLAB_TOKEN": "gl-token",
				"CI_JOB_TOKEN": "job-token",
			},
			expected: http.Header{
				"Authorization": []string{"Bearer gl-token"},
				"Private-Token": []string{"gl-token"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.env {
				_ = os.Setenv(k, v)
			}

			got := ambientCredentials(tt.uri)
			assert.Equal(t, tt.expected, got)
		})
	}
}
