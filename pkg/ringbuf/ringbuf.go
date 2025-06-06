package ringbuf

import "sync"

// RingBuffer[T] is a fixed-size, thread-safe circular buffer.
// When full, Add() will overwrite the oldest element.
type RingBuffer[T any] struct {
    buf   []T
    size  int
    mu    sync.Mutex
    write int // next write index
    count int // number of valid elements currently stored (≤ size)
}

// New creates a RingBuffer that can hold up to `size` elements.
func New[T any](size int) *RingBuffer[T] {
    return &RingBuffer[T]{
        buf:  make([]T, size),
        size: size,
    }
}

// Add inserts item into the buffer, overwriting the oldest entry if full.
func (rb *RingBuffer[T]) Add(item T) {
    rb.mu.Lock()
    defer rb.mu.Unlock()

    rb.buf[rb.write] = item
    rb.write = (rb.write + 1) % rb.size
    if rb.count < rb.size {
        rb.count++
    }
}

// GetAll returns a slice of stored elements in FIFO order (oldest→newest).
func (rb *RingBuffer[T]) GetAll() []T {
    rb.mu.Lock()
    defer rb.mu.Unlock()

    out := make([]T, rb.count)
    // The “start” index is where the oldest element currently lives.
    start := (rb.write + rb.size - rb.count) % rb.size
    for i := 0; i < rb.count; i++ {
        out[i] = rb.buf[(start+i)%rb.size]
    }
    return out
}

// Len returns how many elements are currently stored (0‒size).
func (rb *RingBuffer[T]) Len() int {
    rb.mu.Lock()
    defer rb.mu.Unlock()
    return rb.count
}
