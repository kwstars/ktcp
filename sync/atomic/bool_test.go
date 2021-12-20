package atomic

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBool_isSet(t *testing.T) {
	var b Bool
	assert.Equal(t, false, b.IsSet())
}

func TestBool_setFalse(t *testing.T) {
	var b Bool
	b.SetTrue()
	b.SetFalse()
	assert.Equal(t, false, b.IsSet())
}

func TestBool_setTrue(t *testing.T) {
	var b Bool
	b.SetTrue()
	assert.Equal(t, true, b.IsSet())
}
