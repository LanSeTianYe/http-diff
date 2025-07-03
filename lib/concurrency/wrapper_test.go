package concurrency

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSafeGoWrapper(t *testing.T) {

	wrapper := NewSafeGoWaitGroup()
	assert.NotNil(t, wrapper)

	number := &atomic.Int64{}
	number.Store(0)

	var times int64 = 100

	for i := int64(0); i < times; i++ {
		wrapper.SafeGoWithLogger(func() {
			fmt.Println(1)
			number.Add(1)
		}, func(message any) {})
	}

	wrapper.Wait()

	assert.Equal(t, times, number.Load())
}
