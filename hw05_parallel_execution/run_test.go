package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})
}

func TestRun_EdgeCases(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("m <= 0 means ignore errors", func(t *testing.T) {
		tasks := []Task{
			func() error { return errors.New("fail1") },
			func() error { return errors.New("fail2") },
			func() error { return nil },
		}
		err := Run(tasks, 2, 0) // m <= 0 → игнор ошибок
		require.NoError(t, err, "errors should be ignored when m <= 0")
	})

	t.Run("n <= 0 means run with 1 worker", func(t *testing.T) {
		var counter int32
		tasks := []Task{
			func() error { atomic.AddInt32(&counter, 1); return nil },
		}
		err := Run(tasks, 0, 1)
		require.NoError(t, err)
		require.Equal(t, int32(1), counter, "task should have run")
	})

	t.Run("fewer tasks than workers", func(t *testing.T) {
		var counter int32
		tasks := []Task{
			func() error { atomic.AddInt32(&counter, 1); return nil },
			func() error { atomic.AddInt32(&counter, 1); return nil },
		}
		err := Run(tasks, 10, 1)
		require.NoError(t, err)
		require.Equal(t, int32(2), counter, "all tasks must be executed")
	})

	t.Run("errors less than limit", func(t *testing.T) {
		tasks := []Task{
			func() error { return errors.New("fail1") },
			func() error { return nil },
			func() error { return nil },
		}
		err := Run(tasks, 2, 5) // лимит 5, ошибок меньше
		require.NoError(t, err, "errors less than m should not stop execution")
	})
}
