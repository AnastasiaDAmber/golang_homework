package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic", func(t *testing.T) {
		c := NewCache(3)

		// Заполнение до ёмкости
		require.False(t, c.Set("a", 1))
		require.False(t, c.Set("b", 2))
		require.False(t, c.Set("c", 3))

		// Добавление четвертого — должен вытесниться "a"
		require.False(t, c.Set("d", 4))
		_, ok := c.Get("a")
		require.False(t, ok, "a must be evicted when capacity exceeded")

		_, _ = c.Get("c")
		require.True(t, c.Set("b", 22)) // update + touch

		// Добавление "e" — должен вытесниться "d"
		require.False(t, c.Set("e", 5))
		_, ok = c.Get("d")
		require.False(t, ok, "d must be evicted as LRU")

		// Проверка, что b,c,e на месте и значения корректные
		val, ok := c.Get("b")
		require.True(t, ok)
		require.Equal(t, 22, val)

		val, ok = c.Get("c")
		require.True(t, ok)
		require.Equal(t, 3, val)

		val, ok = c.Get("e")
		require.True(t, ok)
		require.Equal(t, 5, val)
	})

	t.Run("zero capacity cache", func(t *testing.T) {
		c := NewCache(0)

		require.False(t, c.Set("x", 1))
		val, ok := c.Get("x")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("clear cache", func(t *testing.T) {
		c := NewCache(2)
		c.Set("a", 1)
		c.Set("b", 2)

		c.Clear()

		_, ok := c.Get("a")
		require.False(t, ok)

		_, ok = c.Get("b")
		require.False(t, ok)

		require.False(t, c.Set("c", 3))
		val, ok := c.Get("c")
		require.True(t, ok)
		require.Equal(t, 3, val)
	})
}

func TestCacheMultithreading(_ *testing.T) {
	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}
