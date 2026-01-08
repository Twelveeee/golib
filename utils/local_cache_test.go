package utils

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestLocalCache_NewLocalCache(t *testing.T) {
	t.Run("创建新的本地缓存实例", func(t *testing.T) {
		expire := time.Hour
		cache := NewLocalCache(expire)

		if cache == nil {
			t.Fatal("缓存实例不应为 nil")
		}

		if cache.items == nil {
			t.Fatal("缓存 items 不应为 nil")
		}

		if len(cache.items) != 0 {
			t.Errorf("新创建的缓存 items 长度应为 0，实际为 %d", len(cache.items))
		}

		if cache.expire != expire {
			t.Errorf("缓存过期时间应为 %v，实际为 %v", expire, cache.expire)
		}
	})
}

func TestLocalCache_SetAndGet(t *testing.T) {
	t.Run("设置和获取缓存数据", func(t *testing.T) {
		cache := NewLocalCache(time.Hour)
		key := "test_key"
		value := "test_value"

		// 设置缓存
		cache.Set(key, value)

		// 获取缓存
		result, exists := cache.Get(key)

		if !exists {
			t.Error("缓存应存在")
		}

		if result != value {
			t.Errorf("缓存值应为 %v，实际为 %v", value, result)
		}
	})

	t.Run("获取不存在的缓存", func(t *testing.T) {
		cache := NewLocalCache(time.Hour)
		key := "nonexistent_key"

		result, exists := cache.Get(key)

		if exists {
			t.Error("缓存不应存在")
		}

		if result != nil {
			t.Errorf("缓存值应为 nil，实际为 %v", result)
		}
	})
}

func TestLocalCache_Delete(t *testing.T) {
	t.Run("删除缓存数据", func(t *testing.T) {
		cache := NewLocalCache(time.Hour)
		key := "test_key"
		value := "test_value"

		// 设置缓存
		cache.Set(key, value)

		// 确认缓存存在
		if _, exists := cache.Get(key); !exists {
			t.Fatal("缓存应存在")
		}

		// 删除缓存
		cache.Delete(key)

		// 确认缓存已被删除
		result, exists := cache.Get(key)
		if exists {
			t.Error("缓存应已被删除")
		}

		if result != nil {
			t.Errorf("缓存值应为 nil，实际为 %v", result)
		}
	})

	t.Run("删除不存在的缓存", func(t *testing.T) {
		cache := NewLocalCache(time.Hour)
		key := "nonexistent_key"

		// 删除不存在的缓存，不应 panic
		cache.Delete(key)
	})
}

func TestLocalCache_Clear(t *testing.T) {
	t.Run("清空所有缓存", func(t *testing.T) {
		cache := NewLocalCache(time.Hour)

		// 设置多个缓存
		cache.Set("key1", "value1")
		cache.Set("key2", "value2")
		cache.Set("key3", "value3")

		// 确认缓存存在
		if len(cache.items) != 3 {
			t.Fatalf("缓存数量应为 3，实际为 %d", len(cache.items))
		}

		// 清空缓存
		cache.Clear()

		// 确认缓存已被清空
		if len(cache.items) != 0 {
			t.Errorf("缓存数量应为 0，实际为 %d", len(cache.items))
		}

		// 确认所有缓存都不存在
		for _, key := range []string{"key1", "key2", "key3"} {
			if _, exists := cache.Get(key); exists {
				t.Errorf("缓存 %s 应已被清空", key)
			}
		}
	})
}

func TestLocalCache_Expiration(t *testing.T) {
	t.Run("缓存过期测试", func(t *testing.T) {
		// 设置很短的过期时间
		cache := NewLocalCache(10 * time.Millisecond)
		key := "test_key"
		value := "test_value"

		// 设置缓存
		cache.Set(key, value)

		// 立即获取，应存在
		if result, exists := cache.Get(key); !exists || result != value {
			t.Error("缓存应立即存在")
		}

		// 等待过期时间
		time.Sleep(20 * time.Millisecond)

		// 再次获取，应已过期
		result, exists := cache.Get(key)
		if exists {
			t.Error("缓存应已过期")
		}

		if result != nil {
			t.Errorf("缓存值应为 nil，实际为 %v", result)
		}
	})
}

func TestLocalCache_GetOrSet(t *testing.T) {
	t.Run("缓存存在时直接返回", func(t *testing.T) {
		cache := NewLocalCache(time.Hour)
		key := "test_key"
		value := "test_value"

		// 预先设置缓存
		cache.Set(key, value)

		// 调用 GetOrSet
		result, fromCache, err := cache.GetOrSet(key, func() (interface{}, error) {
			return "new_value", nil
		})

		if err != nil {
			t.Errorf("不应有错误，实际为 %v", err)
		}

		if !fromCache {
			t.Error("应从缓存获取")
		}

		if result != value {
			t.Errorf("缓存值应为 %v，实际为 %v", value, result)
		}
	})

	t.Run("缓存不存在时执行函数并设置缓存", func(t *testing.T) {
		cache := NewLocalCache(time.Hour)
		key := "test_key"
		expectedValue := "new_value"

		// 调用 GetOrSet
		result, fromCache, err := cache.GetOrSet(key, func() (interface{}, error) {
			return expectedValue, nil
		})

		if err != nil {
			t.Errorf("不应有错误，实际为 %v", err)
		}

		if fromCache {
			t.Error("不应从缓存获取")
		}

		if result != expectedValue {
			t.Errorf("缓存值应为 %v，实际为 %v", expectedValue, result)
		}

		// 确认缓存已被设置
		if cachedValue, exists := cache.Get(key); !exists || cachedValue != expectedValue {
			t.Error("缓存应已被设置")
		}
	})

	t.Run("函数执行出错时不设置缓存", func(t *testing.T) {
		cache := NewLocalCache(time.Hour)
		key := "test_key"
		expectedError := errors.New("function error")

		// 调用 GetOrSet，函数返回错误
		result, fromCache, err := cache.GetOrSet(key, func() (interface{}, error) {
			return nil, expectedError
		})

		if err != expectedError {
			t.Errorf("错误应为 %v，实际为 %v", expectedError, err)
		}

		if fromCache {
			t.Error("不应从缓存获取")
		}

		if result != nil {
			t.Errorf("结果应为 nil，实际为 %v", result)
		}

		// 确认缓存未被设置
		if _, exists := cache.Get(key); exists {
			t.Error("缓存不应被设置")
		}
	})
}

func TestLocalCache_ConcurrentAccess(t *testing.T) {
	t.Run("并发访问测试", func(t *testing.T) {
		cache := NewLocalCache(time.Hour)
		key := "concurrent_key"
		value := "concurrent_value"
		var wg sync.WaitGroup
		concurrency := 100

		wg.Add(concurrency)

		// 并发设置缓存
		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()
				cache.Set(key, value)
			}()
		}

		wg.Wait()

		// 并发获取缓存
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()
				result, exists := cache.Get(key)
				if !exists {
					t.Error("缓存应存在")
				}
				if result != value {
					t.Errorf("缓存值应为 %v，实际为 %v", value, result)
				}
			}()
		}

		wg.Wait()
	})
}

func TestLocalCache_GetOrSet_Concurrent(t *testing.T) {
	t.Run("并发 GetOrSet 测试 - singleflight 防止缓存击穿", func(t *testing.T) {
		cache := NewLocalCache(time.Hour)
		key := "singleflight_key"
		expectedValue := "singleflight_value"
		var callCount int
		var mu sync.Mutex
		var wg sync.WaitGroup
		concurrency := 10

		wg.Add(concurrency)

		// 并发调用 GetOrSet
		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()

				result, _, err := cache.GetOrSet(key, func() (interface{}, error) {
					mu.Lock()
					callCount++
					mu.Unlock()

					// 模拟耗时操作
					time.Sleep(10 * time.Millisecond)
					return expectedValue, nil
				})

				if err != nil {
					t.Errorf("不应有错误，实际为 %v", err)
				}

				if result != expectedValue {
					t.Errorf("缓存值应为 %v，实际为 %v", expectedValue, result)
				}
			}()
		}

		wg.Wait()

		// 由于使用了 singleflight，函数应该只被调用一次
		if callCount != 1 {
			t.Errorf("函数调用次数应为 1，实际为 %d", callCount)
		}
	})
}

func TestGenerateCacheKey(t *testing.T) {
	t.Run("生成字符串缓存键", func(t *testing.T) {
		input := "test_string"
		expectedKey := "\"test_string\""

		key, err := GenerateCacheKey(input)

		if err != nil {
			t.Errorf("不应有错误，实际为 %v", err)
		}

		if key != expectedKey {
			t.Errorf("缓存键应为 %s，实际为 %s", expectedKey, key)
		}
	})

	t.Run("生成数字缓存键", func(t *testing.T) {
		input := 123
		expectedKey := "123"

		key, err := GenerateCacheKey(input)

		if err != nil {
			t.Errorf("不应有错误，实际为 %v", err)
		}

		if key != expectedKey {
			t.Errorf("缓存键应为 %s，实际为 %s", expectedKey, key)
		}
	})

	t.Run("生成结构体缓存键", func(t *testing.T) {
		type TestStruct struct {
			Name string
			Age  int
		}

		input := TestStruct{Name: "Alice", Age: 30}
		expectedKey := `{"Name":"Alice","Age":30}`

		key, err := GenerateCacheKey(input)

		if err != nil {
			t.Errorf("不应有错误，实际为 %v", err)
		}

		if key != expectedKey {
			t.Errorf("缓存键应为 %s，实际为 %s", expectedKey, key)
		}
	})

	t.Run("生成无法序列化的数据缓存键", func(t *testing.T) {
		// 函数类型无法被 JSON 序列化
		input := func() {}

		_, err := GenerateCacheKey(input)

		if err == nil {
			t.Error("应有错误")
		}
	})
}
