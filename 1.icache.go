package LRUInterView

import "time"

/*
	1. 支持设定过期时间, 精度.
	2. 支持设定最大内存, 当内存超出时候做出合理的处理。
	3. 支持并发安全
	4. 为简化编程细节，无需实现数据落地。
 */
type ICache interface {
	// size 是一个字符串。支持以下参数： 1KB, 100KB, 1MB, 2MB, 1GB 等
	SetMaxMemory(size string) bool
	// 设置一个缓存项，并且在expire 时间之后过期
	Set(key string, val interface{}, expire time.Duration)
	// 获取一个值
	Get(key string) (interface{}, bool)
	// 删除一个值
	Del(key string) bool
	// 检测一个值 是否存在
	Exists(key string) bool
	// 清空所有值
	Flush() bool
	// 返回所有的key 多少
	Keys() int64
}
