package pool

import (
	"bytes"
	"sync"
)

var (
	// GlobalBytesPool 全局共享的bytes对象池，如果各使用方的内存块大小基本一致，可以减少makeSlice的调用次数。
	GlobalBytesPool = NewBytesPool()
)

// BytesPool 复用 bytes.Buffer 的对象池
type BytesPool interface {
	// Get 一个bytes.Buffer。
	Get() *bytes.Buffer

	// Put 一个bytes.Buffer，Put 内部需要对 Buffer 统一做 Reset。
	Put(*bytes.Buffer)
}

// NewBytesPool 创建BytesPool
func NewBytesPool() BytesPool {
	return newBytesPool()
}

func newBytesPool() *bytesPool {
	return &bytesPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(nil)
			},
		},
	}
}

// bytesPool 简单的bytes.Buffer对象池
type bytesPool struct {
	pool *sync.Pool
}

func (p *bytesPool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

func (p *bytesPool) Put(b *bytes.Buffer) {
	b.Reset()
	p.pool.Put(b)
}
