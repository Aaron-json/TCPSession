package client

import (
	"sync"
)

type bufferPool struct {
	pool *sync.Pool
}

var (
	bufPool bufferPool = bufferPool{
		pool: &sync.Pool{
			New: func() any {
				return make([]byte, READ_BUF_SIZE)
			},
		}}
)

func (bp bufferPool) Get() []byte {
	return bp.pool.Get().([]byte)
}

func (bp bufferPool) Put(buf []byte) {
	bp.pool.Put(buf)
}
