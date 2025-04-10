package fetch

import (
	"context"
	"fmt"
	"github.com/carlmjohnson/requests"
	"github.com/go-logr/logr"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func URL(ctx context.Context, src *url.URL) (string, error) {
	log := logr.FromContextOrDiscard(ctx)

	tmp, err := os.MkdirTemp("", "temp-download-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}
	dst := filepath.Join(tmp, path.Base(src.Path))

	log.V(6).Info("downloading file", "src", src, "dst", dst)

	uri := fmt.Sprintf("%s://%s/%s", src.Scheme, src.Host, strings.TrimPrefix(src.Path, "/"))
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
	if token := os.Getenv("GITHUB_TOKEN"); token != "" && (strings.HasPrefix(uri, "https://github.com/") || strings.HasPrefix(uri, "https://api.github.com/")) {
		headers.Set("Authorization", "Bearer "+token)
		return
	}
	if token := os.Getenv("GITLAB_TOKEN"); token != "" && strings.HasPrefix(uri, "https://gitlab.") {
		headers.Set("PRIVATE-TOKEN", token)
		return
	}
	if token := os.Getenv("CI_JOB_TOKEN"); token != "" && strings.HasPrefix(uri, os.Getenv("CI_SERVER_URL")) {
		headers.Set("JOB-TOKEN", token)
		return
	}
	return
}
