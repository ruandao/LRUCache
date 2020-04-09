package LRUInterView

import (
	"fmt"
	sizeUtil "github.com/docker/go-units"
	"math"
	"reflect"
)

//1KB, 100KB, 1MB, 2MB, 1GB 等
func parseSizeStr(size string) int64 {
	s, err := sizeUtil.FromHumanSize(size)
	if err != nil {
		panic(fmt.Sprintf("not available size: %v", err))
	}
	return s
}

// 唔，下面这个 size 计算函数是拷贝过来的，而且还是有问题的
// https://github.com/golang/go/issues/34561
// unsafe.Sizeof 返回的是变量类型占用的内存大小， 不是变量占用的内存大小
// 我有点感觉是我理解错题目了，作为一个面试题不应该这么麻烦，
// 如果 要缓存的两个变量 引用了相同的地址，这个 内存是否要扣减,
// 但是要扣减的话，肯定也是错的，因为这意味着要遍历缓存中所有的存储变量的指针

func DeepSize(v interface{}) int64 {
	return int64(valueSize(reflect.ValueOf(v), make(map[uintptr]bool)))
}

func valueSize(v reflect.Value, seen map[uintptr]bool) uintptr {
	base := v.Type().Size()
	switch v.Kind() {
	case reflect.Ptr:
		p := v.Pointer()
		if !seen[p] && !v.IsNil() {
			seen[p] = true
			return base + valueSize(v.Elem(), seen)
		}

	case reflect.Slice:
		n := v.Len()
		for i := 0; i < n; i++ {
			base += valueSize(v.Index(i), seen)
		}

		// Account for the parts of the array not covered by this slice.  Since
		// we can't get the values directly, assume they're zeroes. That may be
		// incorrect, in which case we may underestimate.
		if cap := v.Cap(); cap > n {
			base += v.Type().Size() * uintptr(cap-n)
		}

	case reflect.Map:
		// A map m has len(m) / 6.5 buckets, rounded up to a power of two, and
		// a minimum of one bucket. Each bucket is 16 bytes + 8*(keysize + valsize).
		//
		// We can't tell which keys are in which bucket by reflection, however,
		// so here we count the 16-byte header for each bucket, and then just add
		// in the computed key and value sizes.
		nb := uintptr(math.Pow(2, math.Ceil(math.Log(float64(v.Len())/6.5)/math.Log(2))))
		if nb == 0 {
			nb = 1
		}
		base = 16 * nb
		for _, key := range v.MapKeys() {
			base += valueSize(key, seen)
			base += valueSize(v.MapIndex(key), seen)
		}

		// We have nb buckets of 8 slots each, and v.Len() slots are filled.
		// The remaining slots we will assume contain zero key/value pairs.
		zk := v.Type().Key().Size()  // a zero key
		zv := v.Type().Elem().Size() // a zero value
		base += (8*nb - uintptr(v.Len())) * (zk + zv)

	case reflect.Struct:
		// Chase pointer and slice fields and add the size of their members.
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			switch f.Kind() {
			case reflect.Ptr:
				p := f.Pointer()
				if !seen[p] && !f.IsNil() {
					seen[p] = true
					base += valueSize(f.Elem(), seen)
				}
			case reflect.Slice:
				base += valueSize(f, seen)
			}
		}

	case reflect.String:
		return base + uintptr(v.Len())
	}
	return base
}