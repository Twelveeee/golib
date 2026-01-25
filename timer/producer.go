package timer

import (
	"sync"
	"time"
)

// Producer 生产者
type Producer interface {
	// 获取当前值(之前已产生的)
	Get() interface{}

	// 注册
	RegisterCallBack(callBackFunc func(value interface{}))

	// 停止
	Stop()
}

// NewProducer 创建一个具有定时器的生产者
// 如 duration= 5分钟，则每5分钟产生一个新的value
// 新生成的值会通过回调告知使用者，或者也可以通过Get方法读取到
func NewProducer(duration time.Duration, producerFn func() interface{}) Producer {
	p := &producer{
		cron:         NewSimpleCron(duration),
		producerFunc: producerFn,
	}

	// 确保初始化之后就有值，这样可以立即使用Get方法读取到
	_ = p.produce()

	p.cron.AddJob(func() {
		val := p.produce()
		p.fire(val)
	})
	return p
}

type producer struct {
	cron         *SimpleCron
	producerFunc func() interface{}
	callBacks    []func(value interface{})
	mu           sync.Mutex
	lastValue    interface{}
}

// RegisterCallBack 注册回调函数，新生成出的内容，将通知给回调函数
func (p *producer) RegisterCallBack(callBackFunc func(value interface{})) {
	p.mu.Lock()
	p.callBacks = append(p.callBacks, callBackFunc)
	p.mu.Unlock()
}

func (p *producer) fire(val interface{}) {
	p.mu.Lock()
	fns := p.callBacks
	p.mu.Unlock()

	for i := 0; i < len(fns); i++ {
		fns[i](val)
	}
}

func (p *producer) produce() interface{} {
	val := p.producerFunc()
	p.mu.Lock()
	p.lastValue = val
	p.mu.Unlock()
	return val
}

// Get 获取当前值
func (p *producer) Get() interface{} {
	p.mu.Lock()
	val := p.lastValue
	p.mu.Unlock()
	return val
}

// Stop 停止
func (p *producer) Stop() {
	p.cron.Stop()
}

var _ Producer = (*producer)(nil)
