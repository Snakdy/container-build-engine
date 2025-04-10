package cache

import v1 "github.com/google/go-containerregistry/pkg/v1"

type Cache interface {
	Put(hash v1.Hash, layer v1.Layer, compressed bool) (v1.Layer, error)
	Get(hash v1.Hash) (v1.Layer, error)
	Delete(hash v1.Hash) error
}
