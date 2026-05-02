package stringutil

// RemoveDuplicates はスライスから重複を除去して出現順を保持したスライスを返します。
func RemoveDuplicates[T comparable](s []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(s))
	for _, v := range s {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}
