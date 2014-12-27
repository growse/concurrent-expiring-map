package cmap

import (
	//	"encoding/json"
	"encoding/binary"
	"sort"
	"strconv"
	"testing"
)

type Animal struct {
	name string
}

func (a *Animal) MarshalBinary() (data []byte, err error) {
	return []byte(a.name), nil
}

func (a *Animal) UnmarshalBinary(data []byte) error {
	a.name = string(data)
	return nil
}

func TestMapCreation(t *testing.T) {
	m := New()
	if m == nil {
		t.Error("map is null.")
	}

	if m.Count() != 0 {
		t.Error("new map should be empty.")
	}
}

func TestInsert(t *testing.T) {
	m := New()
	elephant := Animal{"elephant"}
	monkey := Animal{"monkey"}
	bytes, _ := elephant.MarshalBinary()
	m.Set("elephant", bytes)
	bytes, _ = monkey.MarshalBinary()
	m.Set("monkey", bytes)

	if m.Count() != 2 {
		t.Error("map should contain exactly two elements.")
	}
}

func TestGet(t *testing.T) {
	m := New()

	// Get a missing element.
	val, ok := m.Get("Money")

	if ok == true {
		t.Error("ok should be false when item is missing from map.")
	}

	if len(val) != 0 {
		t.Errorf("Missing values should return as null. Instead got %v", val)
	}

	elephant := Animal{"elephant"}
	bytes, _ := elephant.MarshalBinary()
	m.Set("elephant", bytes)

	// Retrieve inserted element.

	tmp, ok := m.Get("elephant")
	elephant.UnmarshalBinary(tmp)

	if ok == false {
		t.Error("ok should be true for item stored within the map.")
	}

	if &elephant == nil {
		t.Error("expecting an element, not null.")
	}

	if elephant.name != "elephant" {
		t.Error("item was modified.")
	}
}

func TestHas(t *testing.T) {
	m := New()

	// Get a missing element.
	if m.Has("Money") == true {
		t.Error("element shouldn't exist")
	}

	elephant := Animal{"elephant"}
	bytes, _ := elephant.MarshalBinary()
	m.Set("elephant", bytes)

	if m.Has("elephant") == false {
		t.Error("element exists, expecting Has to return True.")
	}
}

func TestRemove(t *testing.T) {
	m := New()

	monkey := Animal{"monkey"}
	bytes, _ := monkey.MarshalBinary()
	m.Set("monkey", bytes)

	m.Remove("monkey")

	if m.Count() != 0 {
		t.Error("Expecting count to be zero once item was removed.")
	}

	temp, ok := m.Get("monkey")

	if ok != false {
		t.Error("Expecting ok to be false for missing items.")
	}

	if temp != nil {
		t.Error("Expecting item to be nil after its removal.")
	}

	// Remove a none existing element.
	m.Remove("noone")
}

func TestFlush(t *testing.T) {
	m := New()
	monkey := Animal{"monkey"}
	bytes, _ := monkey.MarshalBinary()
	m.Set("monkey", bytes)

	m.Flush()

	if m.Count() != 0 {
		t.Errorf("Expecting count to be zero once flushed: %v", m.Count())
	}
}

func TestCount(t *testing.T) {
	m := New()
	for i := 0; i < 100; i++ {
		bytes, _ := (&Animal{strconv.Itoa(i)}).MarshalBinary()
		m.Set(strconv.Itoa(i), bytes)
	}

	if m.Count() != 100 {
		t.Error("Expecting 100 element within map.")
	}
}

func TestIsEmpty(t *testing.T) {
	m := New()

	if m.IsEmpty() == false {
		t.Error("new map should be empty")
	}
	bytes, _ := (&Animal{"elephant"}).MarshalBinary()
	m.Set("elephant", bytes)

	if m.IsEmpty() != false {
		t.Error("map shouldn't be empty.")
	}
}

func TestIterator(t *testing.T) {
	m := New()

	// Insert 100 elements.
	for i := 0; i < 100; i++ {
		bytes, _ := (&Animal{strconv.Itoa(i)}).MarshalBinary()
		m.Set(strconv.Itoa(i), bytes)
	}

	counter := 0
	// Iterate over elements.
	for item := range m.Iter() {
		val := item.Val

		if val == nil {
			t.Error("Expecting an object.")
		}
		counter++
	}

	if counter != 100 {
		t.Error("We should have counted 100 elements.")
	}
}

func TestBufferedIterator(t *testing.T) {
	m := New()

	// Insert 100 elements.
	for i := 0; i < 100; i++ {
		bytes, _ := (&Animal{strconv.Itoa(i)}).MarshalBinary()
		m.Set(strconv.Itoa(i), bytes)
	}

	counter := 0
	// Iterate over elements.
	for item := range m.IterBuffered() {
		val := item.Val

		if val == nil {
			t.Error("Expecting an object.")
		}
		counter++
	}

	if counter != 100 {
		t.Error("We should have counted 100 elements.")
	}
}

func TestConcurrent(t *testing.T) {
	m := New()
	ch := make(chan int)
	const iterations = 1000
	var a [iterations]int

	// Using go routines insert 1000 ints into our map.
	go func() {
		for i := 0; i < iterations/2; i++ {
			// Add item to map.
			bytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(bytes, uint32(i))
			m.Set(strconv.Itoa(i), bytes)

			// Retrieve item from map.
			val, _ := m.Get(strconv.Itoa(i))

			// Write to channel inserted value.
			ch <- int(binary.LittleEndian.Uint32(val))
		} // Call go routine with current index.
	}()

	go func() {
		for i := iterations / 2; i < iterations; i++ {
			bytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(bytes, uint32(i))
			// Add item to map.
			m.Set(strconv.Itoa(i), bytes)

			// Retrieve item from map.
			val, _ := m.Get(strconv.Itoa(i))

			// Write to channel inserted value.
			ch <- int(binary.LittleEndian.Uint32(val))
		} // Call go routine with current index.
	}()

	// Wait for all go routines to finish.
	counter := 0
	for elem := range ch {
		a[counter] = elem
		counter++
		if counter == iterations {
			break
		}
	}

	// Sorts array, will make is simpler to verify all inserted values we're returned.
	sort.Ints(a[0:iterations])

	// Make sure map contains 1000 elements.
	if m.Count() != iterations {
		t.Error("Expecting 1000 elements.")
	}

	// Make sure all inserted values we're fetched from map.
	for i := 0; i < iterations; i++ {
		if i != a[i] {
			t.Error("missing value", i)
		}
	}
}
