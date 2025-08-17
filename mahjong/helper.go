package mahjong

import (
	"math/rand"
	"sort"
)

// RandomByRates 根据权重随机选择索引
func RandomByRates(rates []int) int {
	sum := 0
	for _, rate := range rates {
		sum += rate
	}
	if sum <= 0 {
		return 0
	}

	value := rand.Intn(sum)
	for i, rate := range rates {
		if value < rate {
			return i
		}
		value -= rate
	}
	return 0
}

// RandomByRatesFloat 根据浮点权重随机选择索引
func RandomByRatesFloat(rates []float64) int {
	sum := 0.0
	for _, rate := range rates {
		sum += rate
	}
	if sum <= 0 {
		return 0
	}

	value := rand.Float64() * sum
	for i, rate := range rates {
		if value < rate {
			return i
		}
		value -= rate
	}
	return 0
}

// HexValue 十六进制字符转数值
func HexValue(hexChar byte) int {
	if hexChar >= '0' && hexChar <= '9' {
		return int(hexChar - '0')
	}
	if hexChar >= 'a' && hexChar <= 'f' {
		return 10 + int(hexChar-'a')
	}
	if hexChar >= 'A' && hexChar <= 'F' {
		return 10 + int(hexChar-'A')
	}
	return 0
}

// HexChar 数值转十六进制字符
func HexChar(value int) byte {
	if value >= 0 && value <= 9 {
		return byte('0' + value)
	}
	if value >= 10 && value <= 15 {
		return byte('A' + (value - 10))
	}
	return '?'
}

// HasElement 检查元素是否在容器中
func HasElement[T comparable](container []T, value T) bool {
	for _, v := range container {
		if v == value {
			return true
		}
	}
	return false
}

// HasKey 检查键是否在map中
func HasKey[K comparable, V any](m map[K]V, key K) bool {
	_, ok := m[key]
	return ok
}

// CountElement 计算元素出现次数
func CountElement[T comparable](container []T, value T) int {
	count := 0
	for _, v := range container {
		if v == value {
			count++
		}
	}
	return count
}

// RemoveElement 移除第一个匹配的元素
func RemoveElement[T comparable](container *[]T, value T) bool {
	for i, v := range *container {
		if v == value {
			*container = append((*container)[:i], (*container)[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveElements 移除指定数量的匹配元素
func RemoveElements[T comparable](container *[]T, value T, count int) int {
	removed := 0
	for i := 0; i < len(*container) && removed < count; {
		if (*container)[i] == value {
			*container = append((*container)[:i], (*container)[i+1:]...)
			removed++
		} else {
			i++
		}
	}
	return removed
}

// RemoveAllElement 移除所有匹配的元素
func RemoveAllElement[T comparable](container *[]T, value T) int {
	count := 0
	for i := 0; i < len(*container); {
		if (*container)[i] == value {
			*container = append((*container)[:i], (*container)[i+1:]...)
			count++
		} else {
			i++
		}
	}
	return count
}

// HasSameKeys 检查两个map的键是否相同
func HasSameKeys[K comparable, V1, V2 any](m1 map[K]V1, m2 map[K]V2) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k := range m1 {
		if _, ok := m2[k]; !ok {
			return false
		}
	}
	return true
}

// HasSameElements 检查两个切片元素是否相同
func HasSameElements[T comparable](v1, v2 []T, reorder bool) bool {
	if len(v1) != len(v2) {
		return false
	}

	if reorder {
		v1Copy := make([]T, len(v1))
		v2Copy := make([]T, len(v2))
		copy(v1Copy, v1)
		copy(v2Copy, v2)

		// 使用字符串表示进行比较排序
		sort.Slice(v1Copy, func(i, j int) bool {
			return ToString(v1Copy[i]) < ToString(v1Copy[j])
		})
		sort.Slice(v2Copy, func(i, j int) bool {
			return ToString(v2Copy[i]) < ToString(v2Copy[j])
		})

		return HasSameElements(v1Copy, v2Copy, false)
	}

	for i := range v1 {
		if v1[i] != v2[i] {
			return false
		}
	}
	return true
}

// HasAnyElement 检查容器中是否包含目标集合中的任一元素
func HasAnyElement[T comparable](container []T, targets []T) bool {
	for _, v := range targets {
		if HasElement(container, v) {
			return true
		}
	}
	return false
}

// HasAllElement 检查容器中是否包含目标集合中的所有元素
func HasAllElement[T comparable](container []T, targets []T) bool {
	for _, v := range targets {
		if !HasElement(container, v) {
			return false
		}
	}
	return true
}

// HasAnyKey 检查map中是否包含目标集合中的任一键
func HasAnyKey[K comparable, V any](m map[K]V, targets []K) bool {
	for _, k := range targets {
		if HasKey(m, k) {
			return true
		}
	}
	return false
}

// HasAllKey 检查map中是否包含目标集合中的所有键
func HasAllKey[K comparable, V any](m map[K]V, targets []K) bool {
	for _, k := range targets {
		if !HasKey(m, k) {
			return false
		}
	}
	return true
}

// SetValueIfNotNull 如果指针不为nil则设置值
func SetValueIfNotNull[T any](p *T, value T) {
	if p != nil {
		*p = value
	}
}
