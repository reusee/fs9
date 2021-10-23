package fs9

import "errors"

var (
	ErrInvalidPath  = errors.New("invalid path")
	ErrInvalidName  = errors.New("invalid name")
	ErrFileNotFound = errors.New("file not found")
	ErrOutOfBounds  = errors.New("out of bounds")
	ErrTypeMismatch = errors.New("type mismatch")
	ErrNameMismatch = errors.New("name mismatch")
	ErrDirNotEmpty  = errors.New("dir not empty")
	ErrClosed       = errors.New("closed")
	ErrNodeNotFound = errors.New("node not found")
)
