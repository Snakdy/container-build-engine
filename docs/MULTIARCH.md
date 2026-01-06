# Multi-arch

CBE can build multi-arch OCI images.
This allows you to build an image that works on both ARM and AMD/x86_64.

## How it works

When enabled, CBE will attempt to download the base image as an OCI image index rather than a standard image.

> Note: this is only supported for OCI-formatted image indexes. Docker V2s2 manifests are not yet supported.

It will extract the appropriate image from the index based on the requested platform (e.g., `linux/amd64`, `linux/arm64`).
It will perform a build like normal and append that back onto the index (replacing the previous image).

It will then push the entire index as normal.

## Getting started

Enabling multi-arch is as simple as setting the `GenerateIndex` option to `true`.

> Since multi-arch requires the base image to be a multi-arch index, we recommend that you make this behaviour opt in rather than the default.
> A good pattern is to enable it if the user explicitly sets the build platform (rather than relying on OS auto-detection).

```go
package main

import "github.com/Snakdy/container-build-engine/pkg/builder"

func main() {
	builder.NewBuilder(ctx, "my-base-image", statements, builder.Options{
		GenerateIndex: true,
	})
}
```
