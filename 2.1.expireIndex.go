package LRUInterView

import "sort"

type CacheTimeItem struct {
	idx        int
	expireTime int64
	key        string
}

// 逆序排列
type CacheTimeItemArr []*CacheTimeItem

func (c CacheTimeItemArr) Len() int {
	return len(c)
}

func (c CacheTimeItemArr) Less(i, j int) bool {
	return c[i].expireTime > c[j].expireTime
}

func (c CacheTimeItemArr) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
	// 交互 索引顺序
	// 之所以不直接使用 i, j 来做索引是因为 AddOrUpdate，有可能在前半截 重新排序
	c[i].idx, c[j].idx = c[j].idx, c[i].idx
}

type IExpireIndex interface {
	// 添加或者更新某个item
	AddOrUpdate(x CacheTimeItem)
	// pop 出过期时间最近的一个
	Pop() (c CacheTimeItem, exist bool)
	// 删除某个key
	Del(key string)
}

type ExpireIndex struct {
	data CacheTimeItemArr
	m    map[string]*CacheTimeItem
}

func NewExpireIndex() IExpireIndex {
	return &ExpireIndex{
		data: nil,
		m:    make(map[string]*CacheTimeItem),
	}
}

func (e *ExpireIndex) AddOrUpdate(item CacheTimeItem) {
	// 小的在后面，大的在前面
	var needResortArr CacheTimeItemArr
	v, exist := e.m[item.key]
	if exist {
		// 更新之前key 的过期时间
		preExpiredTime := v.expireTime
		v.expireTime = item.expireTime
		if preExpiredTime > v.expireTime {
			needResortArr = e.data[v.idx:]
		} else {
			needResortArr = e.data[:v.idx]
		}
	} else {
		// 添加新key
		item.idx = len(e.data)
		e.data = append(e.data, &item)
		e.m[item.key] = &item
		needResortArr = e.data
	}
	sort.Sort(needResortArr)
}

func (e *ExpireIndex) _Pop() (c CacheTimeItem, exist bool) {
	// 弹出最后一个元素
	arrLen := e.data.Len()
	if arrLen > 0 {
		c = *e.data[arrLen]
		delete(e.m, c.key)
		e.data = e.data[:arrLen]
		return c, true
	}
	return CacheTimeItem{}, false
}

func (e *ExpireIndex) Pop() (c CacheTimeItem, exist bool) {
	return e._Pop()
}

func (e *ExpireIndex) Del(key string) {
	v, exist := e.m[key]
	if exist {
		// 移到数组末尾，然后抛出
		v.expireTime = -1
		sort.Sort(e.data[v.idx:])
		e._Pop()
	}
}
