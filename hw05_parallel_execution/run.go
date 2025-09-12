package hw05parallelexecution

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		n = 1
	}

	// Если m <= 0, то ошибки игнорируются
	//nolint:gosec
	maxErrors := int32(m)
	ignoreErrors := m <= 0

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tasksCh := make(chan Task)
	var wg sync.WaitGroup
	var errsCount int32

	// Запуск воркеров
	startWorkers(ctx, n, tasksCh, &wg, &errsCount, maxErrors, ignoreErrors, cancel)

	// Запуск продюсерво
	go produceTasks(ctx, tasks, tasksCh)

	wg.Wait()

	if !ignoreErrors && atomic.LoadInt32(&errsCount) >= maxErrors {
		return ErrErrorsLimitExceeded
	}

	return nil
}

// Отправка задач в канал до тех пор,пока они не закончатся или контекст не отменён.
func produceTasks(ctx context.Context, tasks []Task, tasksCh chan<- Task) {
	defer close(tasksCh)

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			return
		case tasksCh <- task:
		}
	}
}

// Запуск n воркеров, каждый из которых берёт задачи из канала и выполняет их.
func startWorkers(
	ctx context.Context,
	n int,
	tasksCh <-chan Task,
	wg *sync.WaitGroup,
	errsCount *int32,
	m int32,
	ignoreErrors bool,
	cancel context.CancelFunc,
) {
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-tasksCh:
					if !ok {
						return
					}
					if err := task(); err != nil {
						if !ignoreErrors && atomic.AddInt32(errsCount, 1) >= m {
							cancel()
							return
						}
					}
				}
			}
		}()
	}
}
