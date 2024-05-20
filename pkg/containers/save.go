package containers

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func Save(ctx context.Context, img v1.Image, dst, path string) error {
	log := logr.FromContextOrDiscard(ctx).WithValues("path", path, "dst", dst)
	log.Info("saving image to local file")
	if err := crane.Save(img, dst, path); err != nil {
		log.Error(err, "failed to save image")
		return err
	}
	return nil
}
