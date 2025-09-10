package utils

func JustPtr[T any](v T) *T {
	return &v
}
