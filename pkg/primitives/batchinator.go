package primitives

func Buckets[T any](in []T, size int) [][]T {
	res := make([][]T, 0, len(in)/size+1)
	for i := 0; i < len(in); i += size {
		end := i + size
		if end > len(in) {
			end = len(in)
		}

		res = append(res, in[i:end])
	}

	return res
}
