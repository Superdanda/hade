package utils

// Filter 泛型过滤函数
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item) // 仅添加符合条件的元素
		}
	}
	return result
}
