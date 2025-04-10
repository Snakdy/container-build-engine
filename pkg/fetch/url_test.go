package fetch

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func TestURL(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))

	t.Run("ambient gitlab ci credentials", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Job-Token") != "hunter2" {
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(`{"foo": "bar"}`))
		}))
		t.Cleanup(ts.Close)

		require.NoError(t, os.Setenv("CI_JOB_TOKEN", "hunter2"))
		require.NoError(t, os.Setenv("CI_SERVER_URL", ts.URL))

		uri, err := url.Parse(ts.URL + "/foo.json")
		require.NoError(t, err)

		_, err = URL(ctx, uri)
		require.NoError(t, err)
	})
}
