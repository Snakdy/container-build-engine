package v1

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAny(t *testing.T) {
	ol := OptionsList{
		{
			"foo": "bar",
		},
		{
			"bar": "zoo",
		},
	}
	t.Run("not present", func(t *testing.T) {
		_, err := GetAny[string](ol, "zoo")
		assert.ErrorIs(t, err, ErrNoValue)
	})
	t.Run("present but wrong type", func(t *testing.T) {
		val, err := GetAny[bool](ol, "foo")
		assert.ErrorIs(t, err, ErrWrongType)
		assert.Empty(t, val)
	})
	t.Run("present", func(t *testing.T) {
		val, err := GetAny[string](ol, "foo")
		assert.NoError(t, err)
		assert.EqualValues(t, "bar", val)
	})
}

func TestGetRequired(t *testing.T) {
	t.Run("not present", func(t *testing.T) {
		o := Options{}
		val, err := GetRequired[string](o, "foo")
		assert.Error(t, err)
		assert.Empty(t, val)
	})
	t.Run("present but wrong type", func(t *testing.T) {
		o := Options{"str": true}
		val, err := GetRequired[string](o, "str")
		assert.Error(t, err)
		assert.Empty(t, val)
	})
	t.Run("present", func(t *testing.T) {
		o := Options{"foo": "bar"}
		val, err := GetRequired[string](o, "foo")
		assert.NoError(t, err)
		assert.EqualValues(t, "bar", val)
	})
}

func TestGetOptional(t *testing.T) {
	t.Run("not present", func(t *testing.T) {
		o := Options{}
		val, err := GetOptional[string](o, "foo")
		assert.NoError(t, err)
		assert.Empty(t, val)
	})
	t.Run("present but wrong type", func(t *testing.T) {
		o := Options{"str": true}
		val, err := GetOptional[string](o, "str")
		assert.Error(t, err)
		assert.Empty(t, val)
	})
	t.Run("present", func(t *testing.T) {
		o := Options{"foo": "bar"}
		val, err := GetOptional[string](o, "foo")
		assert.NoError(t, err)
		assert.EqualValues(t, "bar", val)
	})
}
