package containers

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/chainguard-dev/go-apk/pkg/fs"
	"github.com/go-logr/logr"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

const MagicImageScratch = "scratch"
const DefaultPath = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/home/somebody/.local/bin"

type Image struct {
	author     string
	username   string
	env        []string
	baseImage  v1.Image
	entrypoint []string
	cmd        []string
}

// Deprecated
func (ib *Image) Append(ctx context.Context, fs fs.FullFS, platform *v1.Platform) (v1.Image, error) {
	log := logr.FromContextOrDiscard(ctx)

	// create our new layer
	log.Info("containerising filesystem")
	layer, err := NewLayer(ctx, fs, "somebody", platform)
	if err != nil {
		return nil, err
	}

	// convert the base image to OCI format
	if mt, err := ib.baseImage.MediaType(); err == nil {
		log.V(1).Info("detected base image media type", "mediaType", mt)
	}
	baseImage := ib.baseImage

	// append our layer
	layers := []mutate.Addendum{
		{
			MediaType: types.OCILayer,
			Layer:     layer,
			History: v1.History{
				Author:    ib.author,
				CreatedBy: "container-build-engine",
				Created:   v1.Time{},
			},
		},
	}
	withData, err := mutate.Append(baseImage, layers...)
	if err != nil {
		return nil, fmt.Errorf("appending layers: %w", err)
	}
	// grab a copy of the base image's config file, and set
	// our entrypoint and env vars
	cfg, err := withData.ConfigFile()
	if err != nil {
		return nil, err
	}
	cfg = cfg.DeepCopy()

	// copy platform metadata
	cfg.OS = platform.OS
	cfg.Architecture = platform.Architecture
	cfg.OSVersion = platform.OSVersion
	cfg.Variant = platform.Variant
	cfg.OSFeatures = platform.OSFeatures

	// setup other config bits
	cfg.Author = ib.author
	cfg.Config.WorkingDir = filepath.Join("/home", ib.username)
	cfg.Config.User = ib.username

	log.Info("overriding entrypoint", "before", cfg.Config.Entrypoint, "after", ib.entrypoint)
	cfg.Config.Entrypoint = ib.entrypoint
	log.Info("overriding command", "before", cfg.Config.Cmd, "after", ib.cmd)
	cfg.Config.Cmd = ib.cmd

	cfg.Config.Env = ib.env

	var found bool
	for i, e := range cfg.Config.Env {
		if strings.HasPrefix(e, "PATH=") {
			cfg.Config.Env[i] = cfg.Config.Env[i] + fmt.Sprintf(":/home/%s/.local/bin", ib.username)
			found = true
		}
	}
	if !found {
		cfg.Config.Env = append(cfg.Config.Env, "PATH="+DefaultPath)
	}
	if cfg.Config.Labels == nil {
		cfg.Config.Labels = map[string]string{}
	}

	// package everything up
	img, err := mutate.ConfigFile(withData, cfg)
	if err != nil {
		return nil, err
	}
	return img, nil
}
