package timer

import (
	"errors"
	"sync"
	"time"
)

// NewSimpleCron 创建一个定时任务管理器
//
//	参数 duration 用于控制运行间隔
//	如 1*time.Minute 为每分钟运行一次(每分钟到达的时候触发)
//	1*time.Minute 等效于 unix 的 */1
//	5*time.Minute 等效于 unix 的 */5
func NewSimpleCron(duration time.Duration) *SimpleCron {
	sc := &SimpleCron{
		duration: duration,
	}
	_ = sc.start()
	return sc
}

// SimpleCron 一个简单的定时任务管理器
type SimpleCron struct {
	duration time.Duration
	jobs     []func()

	timer *time.Timer
	mu    sync.Mutex

	checkTimeTimer *time.Ticker
	lastTime       int64

	running bool
}

// AddJob 添加任务
// 添加的任务在运行的时候会新启一个 gor 并行的运行
func (sc *SimpleCron) AddJob(f func()) {
	sc.mu.Lock()
	sc.jobs = append(sc.jobs, f)
	sc.mu.Unlock()
}

// Start 启动任务调度
func (sc *SimpleCron) start() error {
	// 特别的，允许传入0，但是将不会运行
	if sc.duration.Nanoseconds() == 0 {
		return nil
	}
	sc.running = true
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.timer != nil {
		return errors.New("already start")
	}

	sc.timer = time.AfterFunc(sc.next(), func() {
		sc.mu.Lock()
		defer sc.mu.Unlock()
		if !sc.running {
			return
		}
		sc.timer.Reset(sc.next())

		for i := 0; i < len(sc.jobs); i++ {
			go sc.jobs[i]()
		}
	})

	sc.checkTimeTimer = time.NewTicker(time.Second)
	sc.lastTime = nowFunc().Unix()

	go func() {
		for range sc.checkTimeTimer.C {
			sc.checkTimeChange()
		}
	}()
	return nil
}

// checkTimeChange 检查系统的时间是否有发生变化
func (sc *SimpleCron) checkTimeChange() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	nowTime := nowFunc().Unix()
	defer func() {
		sc.lastTime = nowTime
	}()

	if !sc.running {
		return
	}

	dur := nowTime - sc.lastTime
	// 检查时间是否有变动，因为定时器是每1秒触发一次，
	// 所以若时间变动超过2秒，认为系统的时间被修改过
	if dur == 1 || dur == 2 {
		return
	}
	sc.timer.Stop()
	sc.timer.Reset(sc.next())
}

// Stop 停止
func (sc *SimpleCron) Stop() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.running = false
	if sc.timer != nil {
		sc.timer.Stop()
	}

	if sc.checkTimeTimer != nil {
		sc.checkTimeTimer.Stop()
	}
}

func (sc *SimpleCron) next() time.Duration {
	// time.Now().UnixNano() 是相对1970年1月1日（UTC时区）经过的纳秒数
	// 计算各时区相对于这个时间所经过的纳秒数 需要加上各时区相对于UTC时区的偏移量
	_, offsetSec := nowFunc().Zone()
	nowLocalTs := nowFunc().UnixNano() + int64(time.Duration(offsetSec)*time.Second)
	next := int64(sc.duration) - nowLocalTs%int64(sc.duration)
	return time.Duration(next)
}
