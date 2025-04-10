package containers

import (
	"context"
	"fmt"
	"github.com/Snakdy/container-build-engine/pkg/containers/cache"
	"github.com/Snakdy/container-build-engine/pkg/oci/empty"
	"github.com/go-logr/logr"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// Drops docker specific properties
// See: https://github.com/opencontainers/image-spec/blob/main/config.md
func toOCIV1Config(config v1.Config) v1.Config {
	return v1.Config{
		User:         config.User,
		ExposedPorts: config.ExposedPorts,
		Env:          config.Env,
		Entrypoint:   config.Entrypoint,
		Cmd:          config.Cmd,
		Volumes:      config.Volumes,
		WorkingDir:   config.WorkingDir,
		Labels:       config.Labels,
		StopSignal:   config.StopSignal,
	}
}

func toOCIV1ConfigFile(cf *v1.ConfigFile) *v1.ConfigFile {
	return &v1.ConfigFile{
		Created:      cf.Created,
		Author:       cf.Author,
		Architecture: cf.Architecture,
		OS:           cf.OS,
		OSVersion:    cf.OSVersion,
		History:      cf.History,
		RootFS:       cf.RootFS,
		Config:       toOCIV1Config(cf.Config),
	}
}

// NormaliseImage mutates the provided v1.Image to be OCI compliant v1.Image.
//
// Check image-spec to see which properties are ported and which are dropped.
// https://github.com/opencontainers/image-spec/blob/main/config.md
func NormaliseImage(ctx context.Context, base v1.Image) (v1.Image, error) {
	log := logr.FromContextOrDiscard(ctx)
	log.V(2).Info("normalising base image - this may take a while if its the first time")
	log.V(3).Info("we do this to make sure that media type between layers is consistent")
	// get the original manifest
	m, err := base.Manifest()
	if err != nil {
		return nil, err
	}
	// convert config
	cfg, err := base.ConfigFile()
	if err != nil {
		return nil, err
	}
	cfg = toOCIV1ConfigFile(cfg)

	layers, err := base.Layers()
	if err != nil {
		return nil, err
	}

	//goland:noinspection GoPreferNilSlice
	newLayers := []v1.Layer{}

	c := cache.NewFilesystemCache(cache.Dir())

	// go through each layer and convert it to
	// OCI format
	for _, layer := range layers {
		diffId, err := layer.DiffID()
		if err != nil {
			return nil, fmt.Errorf("getting diff id: %w", err)
		}
		// check if we have a cached layer
		if l, err := c.Get(diffId); err == nil {
			log.V(4).Info("skipping layer normalisation as we have a cached copy", "diffId", diffId)
			newLayers = append(newLayers, l)
			continue
		}
		mediaType, err := layer.MediaType()
		if err != nil {
			return nil, fmt.Errorf("getting media type: %w", err)
		}
		log.V(4).Info("checking layer", "mediaType", mediaType, "diffId", diffId)
		switch mediaType {
		case types.DockerLayer:
			layer, err = tarball.LayerFromOpener(layer.Compressed, tarball.WithMediaType(types.OCILayer))
			if err != nil {
				return nil, fmt.Errorf("building layer: %w", err)
			}
			if l, err := c.Put(diffId, layer, true); err == nil {
				layer = l
			}
		case types.DockerUncompressedLayer:
			layer, err = tarball.LayerFromOpener(layer.Uncompressed, tarball.WithMediaType(types.OCIUncompressedLayer))
			if err != nil {
				return nil, fmt.Errorf("building layer: %w", err)
			}
			if l, err := c.Put(diffId, layer, false); err == nil {
				layer = l
			}
		}
		newLayers = append(newLayers, layer)
	}

	log.V(4).Info("appending mutated layers to empty OCI image")
	base, err = mutate.AppendLayers(empty.Image, newLayers...)
	if err != nil {
		return nil, err
	}

	base = mutate.MediaType(base, types.OCIManifestSchema1)
	base = mutate.ConfigMediaType(base, types.OCIConfigJSON)
	base = mutate.Annotations(base, m.Annotations).(v1.Image)
	base, err = mutate.ConfigFile(base, cfg)
	if err != nil {
		return nil, err
	}
	log.V(3).Info("successfully normalised base image")
	return base, nil
}
