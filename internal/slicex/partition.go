package slicex

func Partition[A any](input []A, cb func(A) bool) (trueSet []A, falseSet []A) {
	for _, element := range input {
		if cb(element) {
			trueSet = append(trueSet, element)
		} else {
			falseSet = append(falseSet, element)
		}
	}
	return
}
