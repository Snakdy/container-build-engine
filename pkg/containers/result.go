package containers

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type Result interface {
	MediaType() (types.MediaType, error)
	Size() (int64, error)
	Digest() (v1.Hash, error)
	RawManifest() ([]byte, error)
}

var _ Result = (v1.Image)(nil)
var _ Result = (v1.ImageIndex)(nil)
