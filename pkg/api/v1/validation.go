package v1

import (
	"fmt"
)

func GetOptional[T any](o Options, key string) (T, error) {
	val, ok := o[key]
	if !ok {
		return *new(T), nil
	}
	return val.(T), nil
}

func GetRequired[T any](o Options, key string) (T, error) {
	val, ok := o[key]
	if !ok {
		return *new(T), fmt.Errorf("key '%s' not found", key)
	}
	return val.(T), nil
}
