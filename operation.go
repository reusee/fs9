package fs9

import (
	"io"
	"time"

	"github.com/reusee/e4"
)

type Operation func(
	file *File,
) (
	ret *File,
	err error,
)

func Ensure(
	name string,
	isDir bool,
	create bool,
) Operation {
	return func(
		file *File,
	) (*File, error) {
		if file != nil {
			if file.Name != name {
				panic("impossible")
			}
			if file.IsDir != isDir {
				return nil, ErrTypeMismatch
			}
			return file, nil
		}
		if create {
			return &File{
				Name:    name,
				IsDir:   isDir,
				ModTime: time.Now(),
			}, nil
		}
		return nil, ErrFileNotFound
	}
}

func Write(
	offset int64,
	from io.Reader,
	bytesWritten *int,
) Operation {
	return func(file *File) (*File, error) {
		if offset > file.Size {
			return nil, we.With(
				e4.NewInfo("file size is %d, cannot write at %d", file.Size, offset),
			)(ErrOutOfBounds)
		}
		existed := file.Bytes[:offset]
		content, err := io.ReadAll(from)
		if err != nil {
			return nil, err
		}
		if bytesWritten != nil {
			*bytesWritten = len(content)
		}
		content = append(existed, content...)
		newFile := *file
		newFile.Size = int64(len(content))
		newFile.ModTime = time.Now()
		newFile.Bytes = content
		return &newFile, nil
	}
}

func Read(
	offset int64,
	length int64,
	to io.Writer,
	bytesRead *int,
	eof *bool,
) Operation {
	return func(file *File) (*File, error) {
		end := offset + length
		if end > file.Size {
			end = file.Size
		}
		if end == file.Size && eof != nil {
			*eof = true
		}
		n, err := to.Write(file.Bytes[offset:end])
		if err != nil {
			return nil, err
		}
		if bytesRead != nil {
			*bytesRead = n
		}
		return file, nil
	}
}
