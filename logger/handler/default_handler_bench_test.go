package handler

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"testing"
	"time"
)

// discardWriter 用于测试的丢弃写入器
type discardWriter struct{}

func (d discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// BenchmarkDefaultHandler_Parallel 并发写入基准测试
func BenchmarkDefaultHandler_Parallel(b *testing.B) {
	handler := NewDefaultHandler(discardWriter{}, slog.LevelInfo)
	logger := slog.New(handler)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.InfoContext(ctx, "test message",
				slog.String("key1", "value1"),
				slog.Int("key2", 123),
				slog.Float64("key3", 3.14),
			)
		}
	})
}

// TestDefaultHandler_Concurrent 并发安全性测试
func TestDefaultHandler_Concurrent(t *testing.T) {
	handler := NewDefaultHandler(discardWriter{}, slog.LevelInfo)
	logger := slog.New(handler)
	ctx := context.Background()

	const (
		goroutines = 100
		iterations = 1000
	)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				logger.InfoContext(ctx, "concurrent test",
					slog.Int("goroutine", id),
					slog.Int("iteration", j),
					slog.String("data", "test data"),
				)
			}
		}(i)
	}

	wg.Wait()
}

// TestDefaultHandler_MemoryLeak 内存泄漏测试
func TestDefaultHandler_MemoryLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory leak test in short mode")
	}

	handler := NewDefaultHandler(discardWriter{}, slog.LevelInfo)
	logger := slog.New(handler)
	ctx := context.Background()

	// 预热
	for i := 0; i < 1000; i++ {
		logger.InfoContext(ctx, "warmup",
			slog.String("key", "value"),
			slog.Int("num", i),
		)
	}

	// 记录初始内存
	var m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// 大量写入
	const iterations = 100000
	for i := 0; i < iterations; i++ {
		logger.InfoContext(ctx, "memory leak test",
			slog.String("key1", "value1"),
			slog.Int("key2", i),
			slog.Float64("key3", 3.14),
		)
	}

	// 记录最终内存
	var m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// 计算内存增长
	allocDiff := m2.TotalAlloc - m1.TotalAlloc
	allocPerOp := allocDiff / iterations

	t.Logf("Total allocations: %d bytes", allocDiff)
	t.Logf("Allocations per operation: %d bytes", allocPerOp)
	t.Logf("Heap objects: %d -> %d (diff: %d)", m1.HeapObjects, m2.HeapObjects, m2.HeapObjects-m1.HeapObjects)

	// 如果每次操作分配超过 1KB，可能存在内存泄漏
	if allocPerOp > 1024 {
		t.Errorf("Potential memory leak: %d bytes allocated per operation", allocPerOp)
	}
}

// TestDefaultHandler_StressTest 压力测试
func TestDefaultHandler_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	handler := NewDefaultHandler(discardWriter{}, slog.LevelInfo)
	logger := slog.New(handler)

	const (
		duration   = 10 * time.Second
		goroutines = 50
	)

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	var (
		totalOps int64
		mu       sync.Mutex
	)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	start := time.Now()

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			ops := 0
			for {
				select {
				case <-ctx.Done():
					mu.Lock()
					totalOps += int64(ops)
					mu.Unlock()
					return
				default:
					logger.InfoContext(ctx, "stress test message",
						slog.Int("goroutine", id),
						slog.Int("operation", ops),
						slog.String("data", "some test data here"),
						slog.Time("timestamp", time.Now()),
					)
					ops++
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	opsPerSec := float64(totalOps) / elapsed.Seconds()

	t.Logf("Stress test completed:")
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Total operations: %d", totalOps)
	t.Logf("  Operations/sec: %.2f", opsPerSec)
	t.Logf("  Goroutines: %d", goroutines)
}

// BenchmarkDefaultHandler_vs_TextHandler 与标准库对比
func BenchmarkDefaultHandler_vs_TextHandler(b *testing.B) {
	b.Run("DefaultHandler", func(b *testing.B) {
		handler := NewDefaultHandler(discardWriter{}, slog.LevelInfo)
		logger := slog.New(handler)
		ctx := context.Background()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logger.InfoContext(ctx, "test message",
				slog.String("key1", "value1"),
				slog.Int("key2", 123),
			)
		}
	})

	b.Run("TextHandler", func(b *testing.B) {
		handler := slog.NewTextHandler(discardWriter{}, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		})
		logger := slog.New(handler)
		ctx := context.Background()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logger.InfoContext(ctx, "test message",
				slog.String("key1", "value1"),
				slog.Int("key2", 123),
			)
		}
	})
}
