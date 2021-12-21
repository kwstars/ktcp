package ksync

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventHasFired(t *testing.T) {
	e := NewEvent()
	assert.Equal(t, false, e.HasFired())
	assert.Equal(t, true, e.Fire())
	assert.Equal(t, false, e.Fire())
	assert.Equal(t, true, e.HasFired())
}

func TestEventDoneChannel(t *testing.T) {
	e := NewEvent()

	select {
	case <-e.Done():
		assert.Fail(t, "e.HasFired() = true; want false")
	default:
		assert.Equal(t, false, e.HasFired())
	}

	assert.Equal(t, true, e.Fire())

	select {
	case <-e.Done():
		assert.Equal(t, true, e.HasFired())
	default:
		assert.Fail(t, "e.HasFired() = false; want true")
	}
}

func TestEventMultipleFires(t *testing.T) {
	e := NewEvent()
	assert.Equal(t, false, e.HasFired())
	assert.Equal(t, true, e.Fire())

	for i := 0; i < 3; i++ {
		assert.Equal(t, true, e.HasFired())
		assert.Equal(t, false, e.Fire())
	}
}
