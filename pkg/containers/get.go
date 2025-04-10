package containers

import (
	"context"
	"fmt"
	"github.com/Snakdy/container-build-engine/pkg/oci/auth"
	"github.com/Snakdy/container-build-engine/pkg/oci/empty"

	"github.com/go-logr/logr"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

const MagicImageScratch = "scratch"

func Get(ctx context.Context, ref string) (v1.Image, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("ref", ref)
	log.Info("getting image")

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

	// normalise the image
	img, err = NormaliseImage(ctx, img)
	if err != nil {
		return nil, fmt.Errorf("normalising %s: %w", ref, err)
	}
	return img, nil
}
