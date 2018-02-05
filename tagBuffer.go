package main

import (
	"errors"
)

const (
	maxBufferSize = 20 // arbitrary size, for demonstration
)

type tagBuffer struct {
	Tags     []string // actual content
	Index    int      // highest index with non-zeroed value
	Capacity int      // should be defined by constant at run time
}

func newTagBuffer() *tagBuffer {
	return &tagBuffer{
		Tags:     make([]string, maxBufferSize),
		Index:    0,
		Capacity: maxBufferSize,
	}
}

func (t *tagBuffer) Add(tag string) error {
	if tag == "" {
		return errors.New("cannot pass empty string")
	}

	if t.Index >= t.Capacity {
		// send query to database asynchronously
		// make tempSlice to disassociate values
		// from buffer
		tempSlice := make([]string, t.Capacity) // optimize this
		copy(tempSlice, t.Tags)
		// fmt.Println("about to save tempSlice: ", tempSlice)
		go saveTags(tempSlice...)

		// reset index -- no need to zero values
		t.Index = 0
	}

	t.Tags[t.Index] = tag
	t.Index++
	return nil
}

// Intended to be used to flush the buffer to
// the database when the writing stream is closed
func (t *tagBuffer) Save() {
	if t.Index == 0 {
		return
	}

	go saveTags(t.Tags[:t.Index]...)
}
