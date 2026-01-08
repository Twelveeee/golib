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
	items  map[string]*CacheItem
	mutex  sync.RWMutex
	expire time.Duration // 缓存过期时间
	group  singleflight.Group
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
	defer lc.mutex.RUnlock()

	if item, exists := lc.items[key]; exists {
		if time.Since(item.Timestamp) < lc.expire {
			return item.Data, true
		}
	}
	return nil, false
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
