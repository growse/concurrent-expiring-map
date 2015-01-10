package cmap

import (
	"hash/fnv"
	"sync"
	"time"
)

var SHARD_COUNT = 32

// TODO: Add Keys function which returns an array of keys for the map.

// A "thread" safe map of type string:Anything.
// To avoid lock bottlenecks this map is dived to several (SHARD_COUNT) map shards.
type ConcurrentMap []*ConcurrentMapShared

// A "thread" safe string to anything map.
type ConcurrentMapShared struct {
	items        map[string]ConcurrentMapItem
	sync.RWMutex // Read Write mutex, guards access to internal map.
}

type ConcurrentMapItem struct {
	Expires time.Time
	Value   []byte
}

// Creates a new concurrent map.
func New() ConcurrentMap {
	m := make(ConcurrentMap, SHARD_COUNT)
	for i := 0; i < SHARD_COUNT; i++ {
		m[i] = &ConcurrentMapShared{items: make(map[string]ConcurrentMapItem)}
	}
	return m
}

// Returns shard under given key
func (m ConcurrentMap) GetShard(key string) *ConcurrentMapShared {
	hasher := fnv.New32()
	hasher.Write([]byte(key))
	return m[int(hasher.Sum32())%SHARD_COUNT]
}

// Sets the given value under the specified key.
func (m *ConcurrentMap) Set(key string, value []byte, expiresAt time.Time) {
	if expiresAt.After(time.Now()) {
		shard := m.GetShard(key)
		newItem := ConcurrentMapItem{Value: make([]byte, len(value)), Expires: expiresAt}
		copy(newItem.Value, value)
		shard.Lock()
		defer shard.Unlock()
		shard.items[key] = newItem
	}
}

// Retrieves an element from map under given key.
func (m ConcurrentMap) Get(key string) ([]byte, bool) {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()

	// Get item from shard.
	val, ok := shard.items[key]
	if val.Expires.Before(time.Now()) {
		return nil, false
	}
	return val.Value, ok
}

// Returns the number of elements within the map.
func (m ConcurrentMap) Count() int {
	count := 0
	for i := 0; i < SHARD_COUNT; i++ {
		shard := m[i]
		shard.RLock()
		for _, item := range shard.items {
			if item.Expires.After(time.Now()) {
				count++
			}
		}
		shard.RUnlock()
	}
	return count
}

// Looks up an item under specified key
func (m *ConcurrentMap) Has(key string) bool {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()

	// See if element is within shard.
	val, ok := shard.items[key]
	if ok && val.Expires.Before(time.Now()) {
		return false
	}
	return ok
}

// Removes an element from the map.
func (m *ConcurrentMap) Remove(key string) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	delete(shard.items, key)
}

func (m ConcurrentMap) Flush() {
	for i := 0; i < SHARD_COUNT; i++ {
		m[i] = &ConcurrentMapShared{items: make(map[string]ConcurrentMapItem)}
	}
}

// Checks if map is empty.
func (m *ConcurrentMap) IsEmpty() bool {
	return m.Count() == 0
}

// Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type Tuple struct {
	Key string
	Val ConcurrentMapItem
}

// Returns an iterator which could be used in a for range loop.
func (m ConcurrentMap) Iter() <-chan Tuple {
	ch := make(chan Tuple)
	go func() {
		// Foreach shard.
		for _, shard := range m {
			// Foreach key, value pair.
			shard.RLock()
			for key, val := range shard.items {
				if val.Expires.After(time.Now()) {
					ch <- Tuple{key, val}
				}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}

// Returns a buffered iterator which could be used in a for range loop.
func (m ConcurrentMap) IterBuffered() <-chan Tuple {
	ch := make(chan Tuple, m.Count())
	go func() {
		// Foreach shard.
		for _, shard := range m {
			// Foreach key, value pair.
			shard.RLock()
			for key, val := range shard.items {
				if val.Expires.After(time.Now()) {
					ch <- Tuple{key, val}
				}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}

func ExpireAll() {
	//TODO Need to implement a mechanism to reclaim memory by periodically removing expired keys
}

//Reviles ConcurrentMap "private" variables to json marshal.
/*
func (m ConcurrentMap) MarshalJSON() ([]byte, error) {
	// Create a temporary map, which will hold all item spread across shards.
	tmp := make(map[string]interface{})

	// Insert items to temporary map.
	for item := range m.Iter() {
		tmp[item.Key] = item.Val
	}
	return json.Marshal(tmp)
}*/

// Concurrent map uses Interface{} as its value, therefor JSON Unmarshal
// will probably won't know which to type to unmarshal into, in such case
// we'll end up with a value of type map[string]interface{}, In most cases this isn't
// out value type, this is why we've decided to remove this functionality.

// func (m *ConcurrentMap) UnmarshalJSON(b []byte) (err error) {
// 	// Reverse process of Marshal.

// 	tmp := make(map[string]interface{})

// 	// Unmarshal into a single map.
// 	if err := json.Unmarshal(b, &tmp); err != nil {
// 		return nil
// 	}

// 	// foreach key,value pair in temporary map insert into our concurrent map.
// 	for key, val := range tmp {
// 		m.Set(key, val)
// 	}
// 	return nil
// }
