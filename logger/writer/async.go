package writer

import (
	"io"
	"sync"
	"time"
)

// NewAsync 创建一个异步的writer
//
//	bufSize 异步队列大小
//	timeout 写超时时间，可以为0，若为0将不超时，阻塞写；若设置为>0的值，当writeTo消费比实际写入多，buf满了将丢弃当前数据
//	writeTo 实际写入的writer
func NewAsync(bufSize int, timeout time.Duration, writeTo io.WriteCloser) io.WriteCloser {
	w := &asyncWriter{
		msgs:    make(chan []byte, bufSize),
		timeout: timeout,
		raw:     writeTo,
		done:    make(chan struct{}),
	}
	go w.consumer()
	return w
}

type asyncWriter struct {
	msgs    chan []byte
	closed  bool
	timeout time.Duration

	raw  io.WriteCloser
	done chan struct{}
	mu   sync.Mutex
}

func (a *asyncWriter) consumer() {
	for p := range a.msgs {
		_, _ = a.raw.Write(p)
	}
	a.done <- struct{}{}
}

func (a *asyncWriter) Write(p []byte) (n int, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return 0, io.ErrClosedPipe
	}

	// 复制数据避免 buffer 被复用导致的数据竞争
	buf := make([]byte, len(p))
	copy(buf, p)

	if a.timeout == 0 {
		a.msgs <- buf
		return len(p), nil
	}
	select {
	case a.msgs <- buf:
		return len(p), nil
	case <-time.After(a.timeout):
		return 0, ErrWriteTimeout
	}
}

func (a *asyncWriter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return nil
	}

	close(a.msgs)
	<-a.done

	a.closed = true
	return a.raw.Close()
}

var _ io.WriteCloser = (*asyncWriter)(nil)
