package v1

import (
	"fmt"
	"strconv"
)

func (o Options) GetOptionalString(key string) (string, error) {
	val, ok := o[key]
	if !ok {
		return "", nil
	}
	return val.(string), nil
}

func (o Options) GetString(key string) (string, error) {
	val, ok := o[key]
	if !ok {
		return "", fmt.Errorf("key '%s' not found", key)
	}
	return val.(string), nil
}

func (o Options) GetOptionalBool(key string) (bool, error) {
	val, ok := o[key]
	if !ok {
		return false, nil
	}
	return strconv.ParseBool(val.(string))
}

func (o Options) GetBool(key string) (bool, error) {
	val, ok := o[key]
	if !ok {
		return false, fmt.Errorf("key '%s' not found", key)
	}
	return strconv.ParseBool(val.(string))
}
