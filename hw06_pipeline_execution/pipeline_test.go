package hw06pipelineexecution

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

func TestPipeline(t *testing.T) {
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})
}

func TestAllStageStop(t *testing.T) {
	wg := sync.WaitGroup{}
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		wg.Wait()

		require.Len(t, result, 0)
	})
}

func TestNilDone(t *testing.T) {
	// проверка, что пайплайн работает корректно, если done == nil
	// (канал отмены не передан)
	in := make(Bi)
	stages := []Stage{
		func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					out <- v.(int) * 10
				}
			}()
			return out
		},
	}

	go func() {
		in <- 1
		in <- 2
		in <- 3
		close(in)
	}()

	result := make([]int, 0)
	for v := range ExecutePipeline(in, nil, stages...) {
		result = append(result, v.(int))
	}

	require.Equal(t, []int{10, 20, 30}, result)
}

func TestEmptyStages(t *testing.T) {
	// проверка, что при пустом списке стадий (stages... == nil)
	// пайплайн возвращает входной канал как есть.
	in := make(Bi)

	go func() {
		in <- "a"
		in <- "b"
		close(in)
	}()

	out := make([]string, 0, 2)
	for v := range ExecutePipeline(in, nil) {
		out = append(out, v.(string))
	}

	require.Equal(t, []string{"a", "b"}, out)
}

func TestImmediateDone(t *testing.T) { //  проверка, что при немедленном закрытии done пайплайн ничего не обрабатывает
	in := make(Bi)
	done := make(Bi)
	close(done) // done сразу закрыт

	stages := []Stage{
		func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					out <- v
				}
			}()
			return out
		},
	}

	go func() {
		in <- 1
		in <- 2
		close(in)
	}()

	result := make([]interface{}, 0, 10)
	for v := range ExecutePipeline(in, done, stages...) {
		result = append(result, v)
	}

	require.Empty(t, result)
}

func TestConcurrencyWithoutSleep(t *testing.T) { // проверка, что хотя бы два стейджа обрабатывают данные параллельно
	in := make(Bi)
	done := make(Bi)

	stage1 := func(in In) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			for v := range in {
				out <- v.(int) + 1
			}
		}()
		return out
	}

	stage2 := func(in In) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			for v := range in {
				out <- v.(int) * 2
			}
		}()
		return out
	}

	stages := []Stage{stage1, stage2}

	go func() {
		for i := 0; i < 100; i++ {
			in <- i
		}
		close(in)
	}()

	count := 0
	go func() {
		for range ExecutePipeline(in, done, stages...) {
			count++
		}
	}()

	require.Eventually(t, func() bool {
		return count > 10
	}, time.Second, 10*time.Millisecond)
}
