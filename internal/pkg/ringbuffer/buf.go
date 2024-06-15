package ringbuffer

import (
	"errors"
	"sync"
)

const (
        // DEfAULT SIZE FOR ALL BUFFERS
        // MUST BE A POWER OF 2 FOR THIS IMPLEMENTATION
        BUF_SIZE = 1 << 7
)

// A ring buffer implementation that supports a single reader and multiple writers.
type RingBuffer[T any] struct {
        buf []T
	read uint32 
	write uint32

        rLock sync.Mutex
}

func NewRingBuffer[T any]() *RingBuffer[T] {
        return &RingBuffer[T]{buf: make([]T, BUF_SIZE)}
}

func (rb *RingBuffer[T])Read() (T, error) {
        if rb.size() == 0 {
                return *new(T), errors.New("Buffer is empty")
        }
        res := rb.buf[rb.read & (BUF_SIZE - 1)]

        rb.read++
        return res, nil
}

func (rb *RingBuffer[T]) Write(buf T) error {
        if rb.size() == BUF_SIZE {
                return errors.New("Buffer is full")
        }
        rb.buf[rb.write & (BUF_SIZE - 1)] = buf
        rb.write++
        return nil
}

func (rb *RingBuffer[T])size() uint32 {
        return  rb.write - rb.read
}

