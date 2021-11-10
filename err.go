package fs9

import "errors"

var (
	ErrBadArgument  = errors.New("bad argument")
	ErrCannotLink   = errors.New("cannot link")
	ErrCannotRemove = errors.New("cannot remove")
	ErrClosed       = errors.New("closed")
	ErrDirNotEmpty  = errors.New("dir not empty")
	ErrFileExisted  = errors.New("file existed")
	ErrFileNotFound = errors.New("file not found")
	ErrImmutable    = errors.New("immutable")
	ErrInvalidName  = errors.New("invalid name")
	ErrInvalidPath  = errors.New("invalid path")
	ErrNameMismatch = errors.New("name mismatch")
	ErrNoPermission = errors.New("no permission")
	ErrNodeNotFound = errors.New("node not found")
	ErrOutOfBounds  = errors.New("out of bounds")
	ErrTypeMismatch = errors.New("type mismatch")
)
