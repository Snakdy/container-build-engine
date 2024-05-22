package utils

func CopyMap(src, dst map[string]any) {
	for k, v := range src {
		dst[k] = v
	}
}
