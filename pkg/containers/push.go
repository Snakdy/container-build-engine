package containers

import (
	"context"
	"fmt"
	"time"

	"github.com/Snakdy/container-build-engine/pkg/oci/auth"
	"github.com/go-logr/logr"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Push uploads a v1.Image or v1.ImageIndex to a remote
// registry
func Push(ctx context.Context, img Result, dst string) error {
	log := logr.FromContextOrDiscard(ctx).WithValues("ref", dst)
	log.Info("pushing image")
	start := time.Now()

	// parse what we just pushed
	ref, err := name.ParseReference(dst)
	if err != nil {
		log.Error(err, "failed to parse reference")
		return err
	}
	// push the image
	switch v := img.(type) {
	case v1.Image:
		err = crane.Push(v, dst, crane.WithContext(ctx), crane.WithAuthFromKeychain(auth.KeyChain(auth.Auth{})))
	case v1.ImageIndex:
		err = remote.WriteIndex(ref, v, remote.WithContext(ctx), remote.WithAuthFromKeychain(auth.KeyChain(auth.Auth{})))
	}
	if err != nil {
		log.Error(err, "failed to push image")
		return err
	}
	d, err := img.Digest()
	if err != nil {
		log.Error(err, "failed to read digest")
		return err
	}
	fmt.Println(ref.String() + "@" + d.String())

	log.Info("pushed image", "duration", time.Since(start))
	return nil
}
