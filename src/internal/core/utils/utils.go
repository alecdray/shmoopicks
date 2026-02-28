package utils

func NewPointer[T any](value T) *T {
	return &value
}
