package LRUInterView

import (
	"time"
)

type LRUInterviewCache struct {
	inner innerCache
}

func NewBindoInterviewCache() ICache {
	return &LRUInterviewCache{
		inner: NewInnerCache(),
	}
}

// size 是一个字符串。支持以下参数： 1KB, 100KB, 1MB, 2MB, 1GB 等
func (b *LRUInterviewCache) SetMaxMemory(sizeStr string) bool {
	size := parseSizeStr(sizeStr)
	b.inner.ClearExpire()
	return b.inner.SetMaxMemory(size)
}

// 设置一个缓存项，并且在 expire 时间之后过期
func (b *LRUInterviewCache) Set(key string, val interface{}, expire time.Duration) {
	b.inner.ClearExpire()
	b.inner.Set(key, val, expire)
}

// 获取一个值
func (b *LRUInterviewCache) Get(key string) (interface{}, bool) {
	b.inner.ClearExpire()
	return b.inner.Get(key)
}

// 删除一个值
func (b *LRUInterviewCache) Del(key string) bool {
	return b.inner.Del(key)
}

// 检测一个值 是否存在
func (b *LRUInterviewCache) Exists(key string) bool {
	b.inner.ClearExpire()
	_, ok := b.inner.Get(key)
	return ok
}

// 清空所有值
func (b *LRUInterviewCache) Flush() bool {
	return b.inner.Flush()
}

// 返回所有的key 多少
func (b *LRUInterviewCache) Keys() int64 {
	b.inner.ClearExpire()
	return b.inner.Keys()
}
