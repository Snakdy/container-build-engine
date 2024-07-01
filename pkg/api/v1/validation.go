package v1

import (
	"errors"
	"fmt"
)

// GetAny retrieves the first matching value from a given OptionsList
// and returns ErrNoValue if nothing could be found.
func GetAny[T any](ol OptionsList, key string) (T, error) {
	var zero T
	for _, l := range ol {
		val, err := GetRequired[T](l, key)
		if err != nil {
			if errors.Is(err, ErrNoValue) {
				continue
			}
			return zero, err
		}
		return val, nil
	}
	return zero, ErrNoValue
}

// GetOptional retrieves a value by a given key. It returns
// the zero value of T if the key could not be found.
func GetOptional[T any](o Options, key string) (T, error) {
	var zero T
	val, err := GetRequired[T](o, key)
	if err != nil {
		if errors.Is(err, ErrNoValue) {
			return zero, nil
		}
		return zero, err
	}
	return val, nil
}

// GetRequired retrieves a value by a given key. It returns
// ErrNoValue if the key could not be found.
func GetRequired[T any](o Options, key string) (T, error) {
	var zero T
	val, ok := o[key]
	if !ok {
		return zero, fmt.Errorf("%w: '%s' not found", ErrNoValue, key)
	}
	casted, ok := val.(T)
	if !ok {
		return zero, fmt.Errorf("%w: '%s' is not a '%T'", ErrWrongType, key, val)
	}
	return casted, nil
}
