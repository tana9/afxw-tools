package stringutil

// RemoveDuplicatesBy は key で導出したキーが重複する要素を除去し、出現順を保持したスライスを返します。
func RemoveDuplicatesBy[T any, K comparable](s []T, key func(T) K) []T {
	seen := make(map[K]struct{}, len(s))
	result := make([]T, 0, len(s))
	for _, v := range s {
		k := key(v)
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// RemoveDuplicates はスライスから重複を除去して出現順を保持したスライスを返します。
func RemoveDuplicates[T comparable](s []T) []T {
	return RemoveDuplicatesBy(s, func(v T) T { return v })
}
