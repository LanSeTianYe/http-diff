package concurrency

import (
	"sync"
)

type SafeGoWaitGroup struct {
	wrapper *sync.WaitGroup
}

func NewSafeGoWaitGroup() *SafeGoWaitGroup {
	return &SafeGoWaitGroup{
		wrapper: &sync.WaitGroup{},
	}
}

func (s *SafeGoWaitGroup) Wait() {
	s.wrapper.Wait()
}

func (s *SafeGoWaitGroup) SafeGoWithLogger(f func(), panicWriter func(message any)) {
	s.wrapper.Add(1)

	go func() {
		defer s.wrapper.Done()

		defer func() {
			if err := recover(); err != nil {
				if panicWriter != nil {
					panicWriter(err)
				}
			}
		}()

		// 运行逻辑
		f()
	}()
}
