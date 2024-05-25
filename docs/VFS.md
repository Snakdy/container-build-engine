# Virtual Filesystems

CBE uses a virtual filesystem when running statements so that the statements can perform privileged actions such is creating or modifying `root`-owned files, without actually needing those privileges.

## Filesystem types

CBE supports in-memory and directory (think `chroot`) based virtual filesystems.
These filesystem types come with tradeoffs that are worth considering.

### In-memory

The in-memory filesystem is the default value.
The filesystem is created in-memory and therefore is not influenced by limitations of your actual physical filesystem.

**Pros:**

* No permission issues
* Good performance

**Cons:**

* High memory usage

### Directory

The directory filesystem creates a temporary directory in the `/tmp` directory (or whatever `os.TempDir()` resolves to) and uses that as the root of the filesystem.
It has significantly lower memory usage, but can suffer from limitations introduced by the underlying filesystem.

**Pros:**

* Low memory usage

**Cons:**

* May run into permissions issues
* Worse performance compared to `memfs`

## Setting the filesystem

The `memfs` is used by default, so no action is needed.
To enable the directory filesystem, you can set a switch in the builder options:

```go
package main

import "github.com/Snakdy/container-build-engine/pkg/builder"

func main() {
	builder.NewBuilder(ctx, "my-base-image", statements, builder.Options{
		DirFS: true,
	})
}
```
