package primitives

func Must[T any](f func() (T, error)) T {
	return MustLazy(f)()
}

func MustLazy[T any](f func() (T, error)) func() T {
	return func() T {
		result, err := f()
		if err != nil {
			panic(err)
		}

		return result
	}
}
