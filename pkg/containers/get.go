package containers

import (
	"context"
	"fmt"
	"time"

	"github.com/Snakdy/container-build-engine/pkg/oci/auth"
	"github.com/Snakdy/container-build-engine/pkg/oci/empty"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"

	"github.com/go-logr/logr"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const MagicImageScratch = "scratch"

// Get returns a v1.Image or a v1.ImageIndex depending
// on what the reference points at.
func Get(ctx context.Context, ref string) (Result, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("ref", ref)
	log.Info("pulling image")

	start := time.Now()

	if ref == MagicImageScratch {
		log.V(7).Info("image requested is a scratch image so we don't need to do anything")
		return empty.Image, nil
	}

	remoteRef, err := name.ParseReference(ref)
	if err != nil {
		return nil, fmt.Errorf("parsing name %s: %w", ref, err)
	}

	// fetch the image without actually
	// pulling it
	rmt, err := remote.Get(remoteRef, remote.WithContext(ctx), remote.WithAuthFromKeychain(auth.KeyChain(auth.Auth{})))
	if err != nil {
		return nil, fmt.Errorf("getting %s: %w", ref, err)
	}

	log.Info("getting image", "mediaType", rmt.MediaType)
	if rmt.MediaType == types.OCIImageIndex || rmt.MediaType == types.DockerManifestList {
		idx, err := rmt.ImageIndex()
		if err != nil {
			return nil, fmt.Errorf("getting image as index: %w", err)
		}
		return idx, nil
	}

	img, err := rmt.Image()
	if err != nil {
		return nil, err
	}

	log.Info("pulled image", "duration", time.Since(start))

	// normalise the image
	img, err = NormaliseImage(ctx, img)
	if err != nil {
		return nil, fmt.Errorf("normalising %s: %w", ref, err)
	}
	return img, nil
}

// GetImage is functionally the same as Get but
// returns a v1.Image
func GetImage(ctx context.Context, ref string) (v1.Image, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("ref", ref)
	log.Info("pulling image")

	start := time.Now()

	if ref == MagicImageScratch {
		log.V(7).Info("image requested is a scratch image so we don't need to do anything")
		return empty.Image, nil
	}

	remoteRef, err := name.ParseReference(ref)
	if err != nil {
		return nil, fmt.Errorf("parsing name %s: %w", ref, err)
	}

	// fetch the image without actually
	// pulling it
	rmt, err := remote.Get(remoteRef, remote.WithContext(ctx), remote.WithAuthFromKeychain(auth.KeyChain(auth.Auth{})))
	if err != nil {
		return nil, fmt.Errorf("getting %s: %w", ref, err)
	}

	img, err := rmt.Image()
	if err != nil {
		return nil, err
	}

	log.Info("pulled image", "duration", time.Since(start))

	// normalise the image
	img, err = NormaliseImage(ctx, img)
	if err != nil {
		return nil, fmt.Errorf("normalising %s: %w", ref, err)
	}
	return img, nil
}
