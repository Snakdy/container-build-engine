package pipelines

import (
	cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"
	"github.com/Snakdy/container-build-engine/pkg/pipelines/utils"
	"github.com/go-logr/logr"
	"os"
	"os/exec"
	"time"
)

type Script struct {
	options cbev1.Options
}

func (s *Script) Run(ctx *BuildContext) error {
	log := logr.FromContextOrDiscard(ctx.Context)
	log.V(7).Info("running statement", "options", s.options)

	command, err := cbev1.GetRequired[string](s.options, "command")
	if err != nil {
		return err
	}
	args, err := cbev1.GetRequired[[]string](s.options, "args")
	if err != nil {
		return err
	}

	log.V(9).Info("running script statement", "command", command, "args", args)
	start := time.Now()

	cmd := exec.CommandContext(ctx.Context, command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Error(err, "script execution failed")
		return err
	}

	log.V(6).Info("script execution completed", "duration", time.Since(start))

	return nil
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
