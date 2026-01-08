package gtask

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// TestGoWaitIntegration 测试 Go 和 Wait 的集成功能
// 这个测试方法同时验证 Go 和 Wait 的协同工作
func TestGoWaitIntegration(t *testing.T) {
	// 测试场景1：基本功能验证
	t.Run("BasicFunctionality", func(t *testing.T) {
		g := &Group{}

		// 使用 Go 添加任务
		g.Go(func() error {
			return nil
		})

		// 使用 Wait 等待并验证结果
		successCount, err := g.Wait()
		if successCount != 1 {
			t.Errorf("期望成功任务数为1，但得到%d", successCount)
		}
		if err != nil {
			t.Errorf("期望没有错误，但得到: %v", err)
		}
	})

	// 测试场景2：混合任务类型
	t.Run("MixedTaskTypes", func(t *testing.T) {
		g := &Group{}

		// 添加不同类型的任务
		taskResults := make(chan string, 4)

		// 成功任务
		g.Go(func() error {
			taskResults <- "success1"
			return nil
		})

		// 失败任务
		g.Go(func() error {
			taskResults <- "fail1"
			return errors.New("失败任务1")
		})

		// 另一个成功任务
		g.Go(func() error {
			taskResults <- "success2"
			return nil
		})

		// panic任务
		g.Go(func() error {
			taskResults <- "panic1"
			panic("panic任务")
		})

		// 使用 Wait 等待所有任务完成
		successCount, err := g.Wait()

		// 验证所有任务都执行了
		completedTasks := 0
		timeout := time.After(time.Second)
		allTasksCompleted := false

		for !allTasksCompleted {
			select {
			case <-taskResults:
				completedTasks++
				if completedTasks == 4 {
					allTasksCompleted = true
				}
			case <-timeout:
				t.Errorf("超时：并非所有任务都完成")
				return
			}
		}

		// 验证 Wait 返回结果
		if successCount != 2 {
			t.Errorf("期望成功任务数为2，但得到%d", successCount)
		}
		if err == nil {
			t.Errorf("期望有错误，但得到nil")
		}
		// 验证错误信息包含预期的内容
		errMsg := err.Error()
		if !contains(errMsg, "失败任务1") || !contains(errMsg, "task panic: panic任务") {
			t.Errorf("错误信息不完整，得到: %s", errMsg)
		}
	})

	// 测试场景3：并发控制下的 Go 和 Wait
	t.Run("ConcurrentControl", func(t *testing.T) {
		g := &Group{
			Concurrent: 2,
		}

		var mu sync.Mutex
		runningTasks := 0
		maxConcurrent := 0
		taskStarted := make(chan bool, 5)
		taskCompleted := make(chan bool, 5)

		// 添加多个耗时任务
		for i := 0; i < 5; i++ {
			g.Go(func() error {
				mu.Lock()
				runningTasks++
				if runningTasks > maxConcurrent {
					maxConcurrent = runningTasks
				}
				mu.Unlock()

				taskStarted <- true

				// 模拟任务执行时间
				time.Sleep(50 * time.Millisecond)

				mu.Lock()
				runningTasks--
				mu.Unlock()

				taskCompleted <- true
				return nil
			})
		}

		// 使用 Wait 等待所有任务完成
		successCount, err := g.Wait()

		// 验证并发限制
		if maxConcurrent > g.Concurrent {
			t.Errorf("并发限制失效，最大并发数%d超过限制%d", maxConcurrent, g.Concurrent)
		}

		// 验证所有任务都完成
		if successCount != 5 {
			t.Errorf("期望成功任务数为5，但得到%d", successCount)
		}
		if err != nil {
			t.Errorf("期望没有错误，但得到: %v", err)
		}

		// 确认所有任务都发送了完成信号
		completedCount := 0
		timeout := false
		for i := 0; i < 5; i++ {
			select {
			case <-taskCompleted:
				completedCount++
			case <-time.After(time.Second):
				t.Errorf("等待任务完成超时")
				timeout = true
			}
			if timeout {
				break
			}
		}
		if completedCount != 5 {
			t.Errorf("期望5个任务完成，但只有%d个完成", completedCount)
		}
	})

	// 测试场景4：不允许部分失败的情况
	t.Run("DisallowSomeFail", func(t *testing.T) {
		g := &Group{
			AllowSomeFail: false,
			Concurrent:    1, // 串行执行，确保任务按顺序执行
		}

		// 使用通道来同步任务执行
		firstTaskDone := make(chan bool)
		secondTaskDone := make(chan bool)

		// 添加第一个任务（成功任务）
		g.Go(func() error {
			close(firstTaskDone) // 通知第一个任务已完成
			return nil
		})

		// 等待第一个任务完成
		<-firstTaskDone

		// 添加第二个任务（失败任务）
		g.Go(func() error {
			close(secondTaskDone) // 通知第二个任务已完成
			return errors.New("失败任务")
		})

		// 等待第二个任务完成
		<-secondTaskDone

		// 添加第三个任务（不应该执行）
		taskExecuted := false
		g.Go(func() error {
			taskExecuted = true
			return nil
		})

		// 使用 Wait 等待任务执行完成
		successCount, err := g.Wait()

		// 验证结果
		if successCount != 1 {
			t.Errorf("期望成功任务数为1，但得到%d", successCount)
		}
		if err == nil {
			t.Errorf("期望有错误，但得到nil")
		}
		if err.Error() != "失败任务" {
			t.Errorf("期望错误信息为'失败任务'，但得到: %v", err)
		}

		// 验证第三个任务没有执行
		if taskExecuted {
			t.Errorf("在不允许部分失败的情况下，失败后的任务不应该执行")
		}
	})

	// 测试场景5：空任务组
	t.Run("EmptyGroup", func(t *testing.T) {
		g := &Group{}

		// 不添加任何任务，直接调用 Wait
		successCount, err := g.Wait()

		// 验证结果
		if successCount != 0 {
			t.Errorf("期望空任务组成功任务数为0，但得到%d", successCount)
		}
		if err != nil {
			t.Errorf("期望空任务组没有错误，但得到: %v", err)
		}
	})
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr))))
}

// findSubstring 查找子字符串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
