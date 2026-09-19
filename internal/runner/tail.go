package runner

import "sync"

type TailBuffer struct {
	mu    sync.Mutex
	limit int
	data  []byte
}

func NewTailBuffer(limit int) *TailBuffer {
	return &TailBuffer{limit: limit}
}

func (b *TailBuffer) Write(payload []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.limit <= 0 {
		return len(payload), nil
	}
	if len(payload) >= b.limit {
		b.data = append(b.data[:0], payload[len(payload)-b.limit:]...)
		return len(payload), nil
	}
	overflow := len(b.data) + len(payload) - b.limit
	if overflow > 0 {
		copy(b.data, b.data[overflow:])
		b.data = b.data[:len(b.data)-overflow]
	}
	b.data = append(b.data, payload...)
	return len(payload), nil
}

func (b *TailBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(append([]byte(nil), b.data...))
}
