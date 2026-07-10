package fetch

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/go-logr/logr"
)

func URL(ctx context.Context, src *url.URL) (string, error) {
	log := logr.FromContextOrDiscard(ctx)

	tmp, err := os.MkdirTemp("", "temp-download-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}
	dst := filepath.Join(tmp, path.Base(src.Path))

	log.V(6).Info("downloading file", "src", src, "dst", dst)

	uri := fmt.Sprintf("%s://%s%s", src.Scheme, src.Host, src.EscapedPath())
	err = requests.URL(uri).
		Headers(ambientCredentials(uri)).
		ToFile(dst).
		Fetch(ctx)
	if err != nil {
		_ = os.Remove(dst)
		return "", fmt.Errorf("downloading file: %w", err)
	}
	return dst, nil
}

// ambientCredentials detects credentials from CI systems
// such as GitHub Actions and GitLab CI.
func ambientCredentials(uri string) (headers http.Header) {
	headers = http.Header{}
	if token := cmp.Or(os.Getenv("GITHUB_TOKEN"), os.Getenv("GH_TOKEN")); token != "" && (strings.HasPrefix(uri, "https://github.com/") || strings.HasPrefix(uri, "https://api.github.com/")) {
		headers.Set("Authorization", "Bearer "+token)
		return
	}

	gitlabToken := cmp.Or(os.Getenv("GITLAB_TOKEN"), os.Getenv("GL_TOKEN"))
	jobToken := os.Getenv("CI_JOB_TOKEN")
	serverURL := os.Getenv("CI_SERVER_URL")

	isGitLab := strings.HasPrefix(uri, "https://gitlab.com/") ||
		(serverURL != "" && strings.HasPrefix(uri, serverURL)) ||
		strings.HasPrefix(uri, "https://gitlab.")

	if !isGitLab {
		return
	}
	// set the Authorization header since GitLab seems to use it for more endpoints
	// than just setting the JOB-TOKEN header
	headers.Set("Authorization", "Bearer "+cmp.Or(gitlabToken, jobToken))

	if token := gitlabToken; token != "" {
		headers.Set("PRIVATE-TOKEN", token)
		return
	}
	if token := jobToken; token != "" {
		headers.Set("JOB-TOKEN", token)
		return
	}
	return
}
