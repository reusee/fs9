package fs9

import "errors"

var (
	ErrInvalidPath  = errors.New("invalid path")
	ErrFileNotFound = errors.New("file not found")
	ErrOutOfBounds  = errors.New("out of bounds")
	ErrTypeMismatch = errors.New("type mismatch")
)
