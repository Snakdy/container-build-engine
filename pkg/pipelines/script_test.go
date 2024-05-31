package pipelines

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
	"testing"
)

// interface guard
var _ PipelineStatement = &Script{}

func TestScript_Run(t *testing.T) {
	ctx := logr.NewContext(context.TODO(), testr.NewWithOptions(t, testr.Options{Verbosity: 10}))

	s := &Script{options: map[string]any{
		"command": "ls",
		"args":    []string{"-la", "."},
	}}
	err := s.Run(&BuildContext{
		Context:          ctx,
		WorkingDirectory: "",
	})
	assert.NoError(t, err)
}
