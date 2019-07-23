# Product Name
pool

Pool is another implemention of pool struct in golang.

comparing to sync.Pool in standard library, it has the following features:

    1. it has capacity.
    2. it gurantees that the entity in pool must not be released by gc.
    3. you can pass EntityMaker interface implemention as a maker, which
       means more flexibility and neater code.

# Install

```
go get github.com/lishulong/pool
```

# Usage Example
```
package main

import (
    "log"

    "github.com/lishulong/pool"
)

// ByteSliceMaker implements pool.EntityMaker
type ByteSliceMaker struct {
    size    int
    Counter int
}

func (bsMaker *ByteSliceMaker) Make() (interface{}, error) {
    bsMaker.Counter += 1

    return make([]byte, 0, bsMaker.size), nil
}

func NewByteSliceMaker(size int) *ByteSliceMaker {
    return &ByteSliceMaker{
        size:    size,
        Counter: 0,
    }
}

func main() {
    bsMaker := NewByteSliceMaker(100)
    p := pool.New(bsMaker, 10, pool.WithLock)

    bs1, _ := p.Get()
    log.Printf("first bs addr: %p\n", bs1.([]byte))

    p.Put(bs1)

    bs2, _ := p.Get()
    log.Printf("second bs addr: %p\n", bs2.([]byte))

    log.Printf("allocate cnt: %d\n", bsMaker.Counter)
}
```

for more detailed, see [godoc](https://godoc.org/github.com/lishulong/pool)

# Author

Shulong, Li (李树龙)


# Copyright and License

The MIT License (MIT)

Copyright (c) 2019 Shulong, Li (李树龙)
