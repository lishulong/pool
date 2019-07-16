package pool

import (
	"errors"
	"sync"
)

const (
	WithLock = iota
	WithoutLock
)

// EntityMaker defined the interface to New a entity in pool
type EntityMaker interface {
	Make() (interface{}, error)
}

type Pool struct {
	stoChan  chan interface{} // internally use a channel to store
	Maker    EntityMaker
	Capacity int

	// used to protect Resize from Put and Get
	// without the need of Resize, there is no need of lock,
	// because Pool use channel which is naturely thread-safe.
	lock *sync.Mutex
}

func New(maker EntityMaker, capacity int, mode int) *Pool {
	p := &Pool{
		stoChan:  make(chan interface{}, capacity),
		Maker:    maker,
		Capacity: capacity,
	}

	if mode == WithLock {
		p.lock = &sync.Mutex{}
	}

	return p
}

// Get gets one entity from pool
// if the pool is empty, use EntityMaker to create a new one
func (p *Pool) Get() (interface{}, error) {
	var (
		entity interface{}
		err    error
		closed bool
	)

	if p.lock != nil {
		p.lock.Lock()
		defer p.lock.Unlock()
	}

	select {
	case entity, closed = <-p.stoChan:
		if !closed {
			err = errors.New("pool destroied")
		}
	default:
		entity, err = p.Maker.Make()
	}

	return entity, err
}

// Put puts an entity into pool
// if Put operate on an destroied pool, it will return error.
func (p *Pool) Put(entity interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// runtime.plainError: send on closed channel
			err = r.(error)
		}
	}()

	if p.lock != nil {
		p.lock.Lock()
		defer p.lock.Unlock()
	}

	select {
	case p.stoChan <- entity:
	default:
		err = errors.New("exceed capacity limit")
	}

	return err
}

func (p *Pool) Size() int {
	return len(p.stoChan)
}

func (p *Pool) closeChan() []interface{} {
	close(p.stoChan)

	rest := make([]interface{}, 0)
	for entity := range p.stoChan {
		rest = append(rest, entity)
	}

	return rest

}

// Destroy destroies the pool
// it also returns the rest of entities in pool
func (p *Pool) Destroy() []interface{} {
	if p.lock != nil {
		p.lock.Lock()
		defer p.lock.Unlock()
	}

	return p.closeChan()
}

// Resize resizes the capacity of Pool
// all entities stored in Pool before Resize will be kept,
// if new capacity is big enough for them, they are kept in Pool,
// otherwise, the extra part of them will be returned.
//
// if you want Resize runs concurrently with Get and Put,
// you'd better use WithLock mode, otherwise Get and Put may return error,
// and the bad thing is the error is caused by internall channel, you may not
// know how to handle it.
func (p *Pool) Resize(capacity int) []interface{} {
	if p.lock != nil {
		p.lock.Lock()
		defer p.lock.Unlock()
	}

	if capacity == p.Capacity {
		return []interface{}{}
	}

	stoChan := make(chan interface{}, capacity)

	rest := p.closeChan()
	p.stoChan = stoChan
	p.Capacity = capacity

	num := capacity
	if num > len(rest) {
		num = len(rest)
	}

	for cnt := 0; cnt < num; cnt++ {
		p.stoChan <- rest[cnt]
	}

	return rest[num:]
}
