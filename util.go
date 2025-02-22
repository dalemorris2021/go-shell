package main

func createRange(start int, end int) []int {
	size := end - start
	r := make([]int, size)

	for i := range size {
		r[i] = i + start
	}

	return r
}
