package slicex

func Map[A, B any](input []A, cb func(A) B) []B {
	output := make([]B, len(input))
	for i, entry := range input {
		output[i] = cb(entry)
	}
	return output
}
