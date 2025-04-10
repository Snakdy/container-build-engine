# Caching

To ensure that images work in all environments, base/parent images are normalised.
Essentially, we just convert non-OCI layers (e.g., `dockerv2s2`) to OCI layers.

This can be slow, especially with very large layers.

CBE will attempt to cache normalised layers so that subsequent runs will be able to avoid the trip to the registry to download the layer, and the time spent converting it.

## Configuration

CBE will store the layers on disk and will choose the first available location:

1. `$XDG_CACHE_DIR/cbe`
2. `$HOME/.cache/cbe`
3. `$TMPDIR/cbe`
4. `/tmp/cbe`

If running CBE in CI, make sure that your CI tool saves the cache directory between executions.
For example:

```yaml
my job:
  variables:
    XDG_CACHE_DIR: $CI_PROJECT_DIR/.cache
  script:
    - cbe build # example command
  cache:
    paths:
      - .cache
```
