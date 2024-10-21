package pipelines

import (
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/go-logr/logr"
	"os"
	"os/exec"
	"time"
)

// Script executes an arbitrary script.
// Accepts the following parameters:
//
// 1. "command": command to execute
//
// 2. "args": additional arguments to pass to the command.
type Script struct {
	options cbev1.Options
}

func (s *Script) Run(ctx *BuildContext, _ ...cbev1.Options) (cbev1.Options, error) {
	log := logr.FromContextOrDiscard(ctx.Context)
	log.V(7).Info("running statement", "options", s.options)

	command, err := cbev1.GetRequired[string](s.options, "command")
	if err != nil {
		return cbev1.Options{}, err
	}
	args, err := cbev1.GetRequired[[]string](s.options, "args")
	if err != nil {
		return cbev1.Options{}, err
	}

	log.V(9).Info("running script statement", "command", command, "args", args)
	start := time.Now()

	cmd := exec.CommandContext(ctx.Context, command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Error(err, "script execution failed", "command", command)
		return cbev1.Options{}, err
	}

	log.V(6).Info("script execution completed", "duration", time.Since(start))

	return cbev1.Options{}, nil
}

func (*Script) Name() string {
	return StatementScript
}

func (s *Script) SetOptions(options cbev1.Options) {
	if s.options == nil {
		s.options = map[string]any{}
	}
	utils.CopyMap(options, s.options)
}
