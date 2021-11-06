package fs9

import "io/fs"

func fileChangeMode(mode fs.FileMode) func(*File) error {
	return func(file *File) error {
		file.Mode = mode
		return nil
	}
}

func fileChagneOwner(uid int, gid int) func(*File) error {
	return func(file *File) error {
		if uid != -1 {
			file.UserID = uid
		}
		if gid != -1 {
			file.GroupID = gid
		}
		return nil
	}
}

func fileTruncate(size int64) func(*File) error {
	return func(file *File) error {
		if file.Size == size {
			return nil
		}
		if size < file.Size {
			file.Content = file.Content[:size]
		} else {
			newContent := make([]byte, size)
			copy(newContent, file.Content)
			file.Content = newContent
		}
		file.Size = size
		return nil
	}
}
