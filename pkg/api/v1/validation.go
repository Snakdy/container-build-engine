package v1

import (
	"fmt"
)

func GetOptional[T any](o Options, key string) (T, error) {
	val, ok := o[key]
	if !ok {
		return *new(T), nil
	}
	casted, ok := val.(T)
	if !ok {
		return *new(T), fmt.Errorf("key '%s' is not a %t", key, val)
	}
	return casted, nil
}

func GetRequired[T any](o Options, key string) (T, error) {
	val, ok := o[key]
	if !ok {
		return *new(T), fmt.Errorf("required key '%s' not found", key)
	}
	casted, ok := val.(T)
	if !ok {
		return *new(T), fmt.Errorf("key '%s' is not a '%t'", key, val)
	}
	return casted, nil
}
