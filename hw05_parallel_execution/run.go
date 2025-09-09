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

			for true {
				mx.Lock()
				task := tasks[0]
				if len(tasks) >= 1 {
					tasks = tasks[1:]
				}
				mx.Unlock()

				if task == nil || errCount >= m {
					return
				}

				err := task()
				if err != nil {
					mx.Lock()
					errCount++
					mx.Unlock()
				}

				mx.Lock()
				tcount := len(tasks)
				mx.Unlock()
				if tcount <= 0 || errCount >= m {
					return
				}
			}
		}()
	}
	wg.Wait()

	if errCount >= m {
		return ErrErrorsLimitExceeded
	}
	return nil
}
