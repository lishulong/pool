package pool

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type countMaker struct {
	counter int
}

func (cm *countMaker) Make() (interface{}, error) {
	ret := cm.counter
	cm.counter += 1

	return ret, nil
}

func doTestPool(t *testing.T, pool *Pool, cm *countMaker) {
	require.Equal(t, 10, pool.Capacity, "wrong capacity")
	require.Equal(t, 10, cap(pool.stoChan), "wrong chan capacity")

	// empty pool
	entity, err := pool.Get()
	require.Nil(t, err, "failed to get")
	require.Equal(t, 0, entity.(int), "not equal")
	require.Equal(t, 1, cm.counter, "countMaker is not called")

	err = pool.Put(entity.(int))
	require.Nil(t, err, "failed to put")
	require.Equal(t, 1, pool.Size(), "failed to put entity in chan")

	// pool with one entity
	entity, err = pool.Get()
	require.Nil(t, err, "failed to get")
	require.Equal(t, 0, entity.(int), "not equal")
	require.Equal(t, 1, cm.counter, "countMaker is called")

	for cnt := 0; cnt < 10; cnt++ {
		err = pool.Put(cnt)

		require.Nil(t, err, "failed to put")
	}

	err = pool.Put(11)
	require.NotNil(t, err, "error should occure")
	require.True(t,
		strings.Contains(err.Error(), "exceed capacity limit"),
		"error not right")

	// Resize
	rest := pool.Resize(20)
	require.Empty(t, rest, "not empty return when extend pool")
	require.Equal(t, 10, pool.Size(), "wrong pool size")
	require.Equal(t, 20, pool.Capacity, "wrong pool capacity")

	rest = pool.Resize(5)
	require.Equal(t, 5, len(rest), "wrong return of resize")
	require.Equal(t, 5, pool.Size(), "wrong size of pool")
	require.Equal(t, 5, pool.Capacity, "wrong pool capacity")
	t.Logf("rest: %v", rest)

	entity, err = pool.Get()
	require.Nil(t, err, "failed to get after Resize")
	err = pool.Put(entity)
	require.Nil(t, err, "failed to put after Resize")

	rest = pool.Destroy()
	require.Equal(t, 5, len(rest), "wrong lenght of rest")

	entity, err = pool.Get()
	require.Nil(t, entity, "get entity from destroied pool")
	require.NotNil(t, err, "no error returned")
	require.True(t,
		strings.Contains(err.Error(), "pool destroied"), "error not right")

	err = pool.Put(1)
	require.NotNil(t, err, "send on closed channel, error should not be nil")
	require.True(t,
		strings.Contains(err.Error(), "send on closed channel"),
		"error not right")
}

func TestPool(t *testing.T) {
	cm := &countMaker{counter: 0}
	pool := New(cm, 10, WithLock)
	doTestPool(t, pool, cm)

	cm = &countMaker{counter: 0}
	pool = New(cm, 10, WithoutLock)
	doTestPool(t, pool, cm)
}
