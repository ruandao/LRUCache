package LRUInterView


// 这边只是对 key 做 lru 处理，以便添加key 的时候，快速的找出要驱逐的key
type ILRUKeys interface {
	GetSize() int64
	GetMaxSize() int64
	SetMaxSize(size int64) (evictedKeys []string)
	// 为了添加新 kv， 需要把旧的 之前的kv， 驱逐，以便腾出空间
	// 也可能是更新 旧 key
	AddOrUpdate(key string, size int64) (evictedKeys []string)
	Del(key string)
}

type LRUKeyItem struct {
	pre *LRUKeyItem
	post *LRUKeyItem
	key string
	size int64
}

type LRUKeys struct {
	currentSize int64
	maxSize int64
	heap *LRUKeyItem
	tail *LRUKeyItem
	m map[string]*LRUKeyItem
}

func NewLRUKeys(size int64) ILRUKeys {
	return &LRUKeys{
		currentSize: 0,
		maxSize:     size,
		heap:        nil,
		tail:        nil,
		m:           make(map[string]*LRUKeyItem),
	}
}

func (L *LRUKeys) GetSize() int64 {
	return L.currentSize
}

func (L *LRUKeys) GetMaxSize() int64 {
	return L.maxSize
}

// 驱逐出至少 deltaSize 大小的空间
// 在当前的 useSize 的基础上
// 如果是更新已有的key
// 如果是新增 key
func (L *LRUKeys) _evictedSize(deltaSize int64) (evictedKeys []string) {
	if deltaSize <= 0 {
		return nil
	}
	if deltaSize > L.currentSize {
		panic("not enough memory")
	}
	evictedItem := L.tail
	for evictedItem != nil {
		if deltaSize <= 0 {
			break
		}
		evictedKeys = append(evictedKeys, evictedItem.key)
		deltaSize -= evictedItem.size
		L.currentSize -= evictedItem.size
		evictedItem = evictedItem.pre
	}
	L.tail = evictedItem
	L.tail.post = nil
	return evictedKeys
}

func (L *LRUKeys) SetMaxSize(size int64) (evictedKeys []string) {
	if size <= 0 {
		panic("maxSize should large 0")
	}
	if L.maxSize < size {
		// 不需要驱逐
		return nil
	}
	// 计算容量差，找出要驱逐的keys
	deltaSize := L.currentSize - size
	evictedKeys = L._evictedSize(deltaSize)
	return evictedKeys
}

func (L *LRUKeys) _Update(key string, size int64) (evictedKeys []string) {
	// 更新比较麻烦，所以我们宁愿说先删除，然后再添加进来
	L._Del(key)
	return L._Add(key, size)
}

func (L *LRUKeys) _Del(key string) {
	v, exist := L.m[key]
	if !exist {
		return
	}
	delete(L.m, key)
	L.currentSize -= v.size

	// 处理麻烦的链
	pre := v.pre
	post := v.post
	if pre != nil {
		pre.post = post
	}
	if post != nil {
		post.pre = pre
	}

	if L.tail == v {
		// 尾指针
		L.tail = pre
	}
	if L.heap == v {
		// 头部指针
		L.heap = post
	}
}

func (L *LRUKeys) _Add(key string, size int64) (evictedKeys []string) {
	// 添加一个新的key 进来
	// 这里不需要考虑说key之前已经存在，如果已经存在，不应该调用到这里
	deltaSize := size - (L.maxSize - L.currentSize)
	evictedKeys = L._evictedSize(deltaSize)
	item := &LRUKeyItem{
		pre:  nil,
		post: L.heap,
		key:  key,
		size: size,
	}
	L.m[key] = item
	L.heap.pre = item
	L.heap = item
	L.currentSize = L.currentSize + size
	return evictedKeys
}

func (L *LRUKeys) AddOrUpdate(key string, size int64) (evictedKeys []string) {
	_, exist := L.m[key]
	if exist {
		return L._Update(key, size)
	} else {
		return L._Add(key, size)
	}
}

func (L *LRUKeys) Del(key string) {
	L._Del(key)
}
