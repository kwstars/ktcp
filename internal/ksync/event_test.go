/*
 *
 * Copyright 2018 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
