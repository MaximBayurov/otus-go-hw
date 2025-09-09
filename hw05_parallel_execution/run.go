package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if m <= 0 {
		return ErrErrorsLimitExceeded
	}
	if len(tasks) == 0 || n <= 0 {
		return nil
	}

	errCount := 0
	mx := sync.Mutex{}
	wg := sync.WaitGroup{}
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				mx.Lock()
				task := tasks[0]
				if len(tasks) >= 1 {
					tasks = tasks[1:]
				}

				if task == nil || errCount >= m {
					mx.Unlock()
					return
				}
				mx.Unlock()

				err := task()
				if err != nil {
					mx.Lock()
					errCount++
					mx.Unlock()
				}

				mx.Lock()
				tcount := len(tasks)
				if tcount <= 0 || errCount >= m {
					mx.Unlock()
					return
				}
				mx.Unlock()
			}
		}()
	}
	wg.Wait()

	if errCount >= m {
		return ErrErrorsLimitExceeded
	}
	return nil
}
