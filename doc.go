package pool

// Pool is another implemention of pool struct in golang.
// comparing to sync.Pool in standard library, it has the following features:
//
//     1. it has capacity.
//     2. it gurantees that the entity in pool must not be released by gc.
//     3. you can pass EntityMaker interface implemention as a maker, which
//        means more flexibility and neater code.
//
// Pool supplies Get, Put, Destroy, Size and Resize methods.
// Pool use channel internally, so without Resize, Get, Put Destroy are all
// naturally thread-safe.
//
// But in Resize, we have to make a new channel of specified buffer size,
// close old channel and copy the entity from old channel to new channel.
// this process must be mutual exclusion from Put and Get, otherwise, they
// may return error caused by internal channel which is really hard for
// user to handle.
//
// So if you need Resize to run concurrently with other APIs,
// use WithLock mode, otherwise use WithoutLock mode.
