package utils

import "sync"

var panicHandler func(info interface{})

// SetPanicHandler 统一将goroutine的panic管理起来
func SetPanicHandler(hd func(info interface{})) {
	panicHandler = hd
}

// SafeGo 安全的使用goroutine
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				if panicHandler != nil {
					panicHandler(err)
				}
			}
		}()
		fn()
	}()
}

// CallbackGo 安全使用go的同时，额外的保证在goroutine执行结束后调用回调函数，即使panic也会出发回调
func CallbackGo(fn func(), callback func()) {
	go func() {
		defer func() {
			callback()
			if err := recover(); err != nil {
				if panicHandler != nil {
					panicHandler(err)
				}
			}
		}()
		fn()
	}()
}

type OnceErr struct {
	err  error
	once sync.Once
}

// SetError ...
func (n *OnceErr) SetError(err error) {
	n.once.Do(func() {
		n.err = err
	})
}

func (n *OnceErr) Error() error {
	return n.err
}
