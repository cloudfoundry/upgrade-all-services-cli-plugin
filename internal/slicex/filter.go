package slicex

func Filter[A any](a []A, cb func(A) bool) (result []A) {
	for _, v := range a {
		if cb(v) {
			result = append(result, v)
		}
	}
	return
}
