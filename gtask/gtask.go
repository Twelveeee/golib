package gtask

import (
	"fmt"
	"sync"
)

// Group 表示一个并发任务组
type Group struct {
	Concurrent    int  // 最大并发数，0表示不限制
	AllowSomeFail bool // 是否允许部分失败

	wg           sync.WaitGroup // 用于等待所有任务完成
	semaphore    chan struct{}  // 用于控制并发数的信号量
	mu           sync.Mutex     // 互斥锁，保护共享状态
	errors       []error        // 收集所有错误
	successCount int            // 成功任务计数
	totalTasks   int            // 总任务数
	once         sync.Once      // 用于一次性初始化资源
}

// Go 添加一个任务到任务组中
func (g *Group) Go(task func() error) {
	// 一次性初始化资源
	g.once.Do(func() {
		g.errors = make([]error, 0)
		// 初始化信号量通道
		if g.Concurrent > 0 {
			g.semaphore = make(chan struct{}, g.Concurrent)
		}
	})

	// 如果不允许部分失败，检查是否已经有失败
	if !g.AllowSomeFail && g.getHasFailed() {
		return
	}

	g.addTotalTasks()
	g.wg.Add(1)

	// 不做并发控制
	if g.Concurrent == 0 {
		go g.runTask(task)
		return
	}

	// 使用信号量控制并发数
	g.semaphore <- struct{}{}
	go func() {
		defer func() { <-g.semaphore }()
		g.runTask(task)
	}()
}

// Wait 等待所有任务完成，返回是否全部成功和错误信息
func (g *Group) Wait() (int, error) {
	g.wg.Wait()

	successCount, _, errors := g.getStats()

	if len(errors) == 0 {
		return successCount, nil
	}

	if g.AllowSomeFail {
		return successCount, g.joinErrors()
	}

	return successCount, g.joinErrors()
}

// addError 添加错误到错误列表
func (g *Group) addError(err error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.errors = append(g.errors, err)
}

// addTotalTasks 增加总任务数
func (g *Group) addTotalTasks() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.totalTasks++
}

// hasFailed 检查是否已经有任务失败
func (g *Group) getHasFailed() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.errors) > 0
}

// addSuccessCount 增加成功计数
func (g *Group) addSuccessCount() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.successCount++
}

// runTask 执行单个任务，包含 recover 机制
func (g *Group) runTask(task func() error) {
	defer g.wg.Done()

	defer func() {
		if r := recover(); r != nil {
			g.addError(fmt.Errorf("task panic: %v", r))
		}
	}()

	err := task()
	if err != nil {
		g.addError(err)
		return
	}

	g.addSuccessCount()
}

// joinErrors 将多个错误拼接成一个错误
func (g *Group) joinErrors() error {
	if len(g.errors) == 0 {
		return nil
	}

	var errMsg string
	for _, err := range g.errors {
		if errMsg != "" {
			errMsg += "; "
		}
		errMsg += err.Error()
	}
	return fmt.Errorf("%s", errMsg)
}

// getStats 获取统计信息
func (g *Group) getStats() (int, int, []error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.successCount, g.totalTasks, g.errors
}
