package rapide

func removeItem[T any](s []T, i int) []T {
	_ = s[i]
	copy(s[i:], s[i+1:])
	return s[:len(s)-1]
}

func findItem[T comparable](s []T, c T) int {
	for i, v := range s {
		if v == c {
			return i
		}
	}
	return -1
}
