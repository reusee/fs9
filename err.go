package fs9

import "errors"

var (
	ErrClosed       = errors.New("closed")
	ErrDirNotEmpty  = errors.New("dir not empty")
	ErrFileExisted  = errors.New("file existed")
	ErrFileNotFound = errors.New("file not found")
	ErrImmutable    = errors.New("immutable")
	ErrInvalidName  = errors.New("invalid name")
	ErrInvalidPath  = errors.New("invalid path")
	ErrNameMismatch = errors.New("name mismatch")
	ErrNodeNotFound = errors.New("node not found")
	ErrOutOfBounds  = errors.New("out of bounds")
	ErrTypeMismatch = errors.New("type mismatch")
	ErrCannotLink   = errors.New("cannot link")
	ErrBadArgument  = errors.New("bad argument")
)
