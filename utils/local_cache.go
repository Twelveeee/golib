package utils

import (
	"encoding/json"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// CacheItem 缓存项结构体
type CacheItem struct {
	Data      interface{} // 缓存数据
	Timestamp time.Time   // 时间戳
}

// LocalCache 本地缓存结构体
type LocalCache struct {
	items map[string]*CacheItem
	mutex sync.RWMutex

	expire time.Duration // 缓存过期时间
	group  singleflight.Group

	cleanupStop chan struct{}
	cleanupDone chan struct{}
	cleanupMu   sync.Mutex
}

// NewLocalCache 创建新的本地缓存实例
func NewLocalCache(expire time.Duration) *LocalCache {
	return &LocalCache{
		items:  make(map[string]*CacheItem),
		expire: expire,
	}
}

// Get 从缓存获取数据
func (lc *LocalCache) Get(key string) (interface{}, bool) {
	lc.mutex.RLock()
	item, exists := lc.items[key]
	if !exists {
		lc.mutex.RUnlock()
		return nil, false
	}
	if time.Since(item.Timestamp) < lc.expire {
		data := item.Data
		lc.mutex.RUnlock()
		return data, true
	}
	lc.mutex.RUnlock()

	// 读锁判断过期后，升级写锁并二次校验后删除，避免竞态误删。
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	item, exists = lc.items[key]
	if !exists {
		return nil, false
	}
	if time.Since(item.Timestamp) >= lc.expire {
		delete(lc.items, key)
		return nil, false
	}

	// 在 RUnlock 与 Lock 之间，可能有其他写入把 key 刷新为最新值；
	// 因此二次校验若发现未过期，应返回最新数据，而不是误判 miss。
	return item.Data, true
}

// Set 设置缓存数据
func (lc *LocalCache) Set(key string, data interface{}) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	lc.items[key] = &CacheItem{
		Data:      data,
		Timestamp: time.Now(),
	}
}

// Delete 删除缓存数据
func (lc *LocalCache) Delete(key string) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	delete(lc.items, key)
}

// Clear 清空所有缓存
func (lc *LocalCache) Clear() {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	lc.items = make(map[string]*CacheItem)
}

// CleanupExpired 批量清理已过期的缓存项，返回清理数量。
func (lc *LocalCache) CleanupExpired() int {
	now := time.Now()
	removed := 0

	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	for key, item := range lc.items {
		if now.Sub(item.Timestamp) >= lc.expire {
			delete(lc.items, key)
			removed++
		}
	}

	return removed
}

// StartAutoCleanup 启动后台定时清理任务。
// interval <= 0 时不启动任何清理任务。
func (lc *LocalCache) StartAutoCleanup(interval time.Duration) {
	if interval <= 0 {
		return
	}

	lc.cleanupMu.Lock()
	defer lc.cleanupMu.Unlock()

	// 已在运行则直接返回
	if lc.cleanupStop != nil {
		return
	}

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})
	lc.cleanupStop = stopCh
	lc.cleanupDone = doneCh

	go func(stop <-chan struct{}, done chan<- struct{}) {
		ticker := time.NewTicker(interval)
		defer func() {
			ticker.Stop()
			close(done)
		}()

		for {
			select {
			case <-ticker.C:
				lc.CleanupExpired()
			case <-stop:
				return
			}
		}
	}(stopCh, doneCh)
}

// StopAutoCleanup 停止后台定时清理任务。
func (lc *LocalCache) StopAutoCleanup() {
	lc.cleanupMu.Lock()
	stopCh := lc.cleanupStop
	doneCh := lc.cleanupDone
	lc.cleanupStop = nil
	lc.cleanupDone = nil
	lc.cleanupMu.Unlock()

	if stopCh == nil {
		return
	}

	close(stopCh)
	<-doneCh
}

// GetOrSet 从缓存获取数据，如果不存在则执行函数获取并设置缓存
func (lc *LocalCache) GetOrSet(key string, fn func() (interface{}, error)) (interface{}, bool, error) {
	if data, exists := lc.Get(key); exists {
		return data, true, nil
	}

	// 使用 singleflight 防止缓存击穿;如果重复执行,只有一个会真正执行,结束后返回值会copy到其他携程
	result, err, _ := lc.group.Do(key, func() (interface{}, error) {
		// 执行函数获取数据
		data, err := fn()
		if err != nil {
			return nil, err
		}

		// 设置缓存
		lc.Set(key, data)
		return data, nil
	})

	return result, false, err
}

// GenerateCacheKey 生成缓存key
func GenerateCacheKey(v interface{}) (string, error) {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
