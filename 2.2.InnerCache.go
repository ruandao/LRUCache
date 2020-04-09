package LRUInterView

import (
	"runtime"
	"sync"
	"time"
)

const defaultMaxMemoSize = 1024 * 1024 // 1 KB
type innerCache struct {
	maxMemoSize   int64
	expireIndex   IExpireIndex           // 最小堆，用于快速确认要过期的key
	evictedIndex  ILRUKeys               // 用于快速找出，哪些key 需要驱逐的
	data          map[string]interface{} // 存放数据
	lock          sync.RWMutex
	lastClearTime int64 // 最后清理的时间
}

func NewInnerCache() innerCache {
	return innerCache{
		expireIndex:   NewExpireIndex(),
		evictedIndex:  NewLRUKeys(defaultMaxMemoSize),
		data:          map[string]interface{}{},
		lastClearTime: time.Now().Unix(),
	}
}

func (i *innerCache) _ClearExpire() {
	nowTime := time.Now().Unix() // Unix 的精度是 秒
	for {
		mostRecentElem, ok := i.expireIndex.Pop()
		if !ok {
			return
		}

		if mostRecentElem.expireTime < nowTime { // 需要清理的元素， 如果刚好等于，则认为不需要清理
			i._Del(mostRecentElem.key)
		} else { // 如果不需要清理，扔回去
			i.expireIndex.AddOrUpdate(mostRecentElem)
			return
		}
	}
}
func (i *innerCache) ClearExpire() {
	nowTime := time.Now().Unix()
	if nowTime > i.lastClearTime {
		// 如果至少过了 一秒, 那么可以清理
		i.lock.Lock()
		defer i.lock.Unlock() // go 1.14 后 defer 的成本很低, 所以就不放在下面了
		nowTime = time.Now().Unix()
		if nowTime > i.lastClearTime { // 如果同一秒内，已经清理过了，那么就不要再次清理
			i._ClearExpire()
			i.lastClearTime = nowTime
		}
	}
}
func (i *innerCache) SetMaxMemory(size int64) bool {

	i.lock.Lock()
	defer i.lock.Unlock()
	// 如果设置的 size 大于系统能提供的最大容量，返回 false
	stat := &runtime.MemStats{}
	runtime.ReadMemStats(stat)
	availMemSize := int64(stat.Frees) + i.maxMemoSize
	if availMemSize  < size {
		return false
	}

	evictedKeys := i.evictedIndex.SetMaxSize(size)
	for _, key := range evictedKeys {
		i._Del(key)
	}
	i.maxMemoSize = size
	return true
}
func (i *innerCache) UseMemoSize() int64 {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.evictedIndex.GetSize()
}

func (i *innerCache) Set(key string, val interface{}, expire time.Duration) {
	item := CacheTimeItem{
		idx:0,
		expireTime: 0,
		key: key,
	}
	lruItem := LRUKeyItem{
		pre: nil,
		post: nil,
		key: key,
		size: 0,
	}
	totalUseSize := DeepSize(item) + DeepSize(lruItem) + DeepSize(val)

	i.lock.Lock()
	defer i.lock.Unlock()
	newExpire := time.Now().Add(expire).Unix() // Unix 的精度是秒
	i.data[key] = val
	cacheTimeItem := CacheTimeItem{key: key, expireTime: newExpire}
	i.expireIndex.AddOrUpdate(cacheTimeItem)
	evictedKeys := i.evictedIndex.AddOrUpdate(key, totalUseSize)
	for _, key := range evictedKeys {
		i._Del(key)
	}

}

func (i *innerCache) Get(key string) (interface{}, bool) {
	i.lock.RLock()
	defer i.lock.RUnlock()
	v, ok := i.data[key]
	return v, ok
}

func (i *innerCache) _Del(key string) bool {
	_, exist := i.data[key]
	if !exist {
		return false
	}
	i.expireIndex.Del(key)
	i.evictedIndex.Del(key)
	// 扣减 内存使用计数
	delete(i.data, key)
	return true
}
func (i *innerCache) Del(key string) bool {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i._Del(key)
}

// 不懂为什么这里需要返回 bool 值
// 希望解释下
func (i *innerCache) Flush() bool {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.evictedIndex = NewLRUKeys(i.maxMemoSize)
	i.expireIndex = NewExpireIndex()
	i.data = make(map[string]interface{})
	i.lastClearTime = time.Now().Unix()
	return true
}

func (i *innerCache) Keys() int64 {
	i.lock.RLock()
	defer i.lock.RUnlock()
	n := len(i.data)
	return int64(n)
}
