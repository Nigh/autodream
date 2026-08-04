package frame

import (
	"sync"
	"sync/atomic"
	"time"
)

// Pool recycles Frame backing buffers to reduce allocations.
type Pool struct {
	mu   sync.Mutex
	free []*Frame
	seq  atomic.Uint64
}

// Acquire returns a frame with a buffer sized for w*h*bpp. Data is zeroed length but capacity set.
func (p *Pool) Acquire(w, h int, fmt PixelFormat) *Frame {
	bpp := fmt.BytesPerPixel()
	need := w * h * bpp
	p.mu.Lock()
	for i, f := range p.free {
		if cap(f.Data) >= need && f.PixelFormat == fmt {
			p.free[i] = p.free[len(p.free)-1]
			p.free = p.free[:len(p.free)-1]
			p.mu.Unlock()
			f.FrameID = p.seq.Add(1)
			f.Timestamp = time.Now().UTC()
			f.Width = w
			f.Height = h
			f.PixelFormat = fmt
			f.Data = f.Data[:need]
			return f
		}
	}
	p.mu.Unlock()
	return &Frame{
		FrameID:     p.seq.Add(1),
		Timestamp:   time.Now().UTC(),
		Width:       w,
		Height:      h,
		PixelFormat: fmt,
		Data:        make([]byte, need),
	}
}

// Release returns a frame to the pool. The frame must not be used after Release.
func (p *Pool) Release(f *Frame) {
	if f == nil {
		return
	}
	f.Data = f.Data[:0]
	f.Width = 0
	f.Height = 0
	p.mu.Lock()
	p.free = append(p.free, f)
	p.mu.Unlock()
}
