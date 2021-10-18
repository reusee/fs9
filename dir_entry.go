package fs9

import (
	"sort"
	"strings"

	"github.com/reusee/e4"
)

type DirEntry struct {
	File       *File
	DirEntries *DirEntries
}

type DirEntries []DirEntry

func (d DirEntries) MinName() string {
	if len(d) == 0 {
		return ""
	}
	entry := d[0]
	if entry.File != nil {
		return entry.File.Name
	} else if entry.DirEntries != nil {
		return entry.DirEntries.MinName()
	}
	panic("impossible")
}

func (d DirEntries) Apply(path []string, op Operation) (newEntries *DirEntries, err error) {
	we := we.With(
		e4.NewInfo("path: %s", strings.Join(path, "/")),
	)

	if len(path) == 0 {
		return nil, we(ErrFileNotFound)
	}

	// name
	name := path[0]
	if name == "" || name == "." || name == ".." {
		return nil, we(ErrInvalidName)
	}

	// descend
	i := sort.Search(len(d), func(i int) bool {
		entry := d[i]
		if entry.File != nil {
			return entry.File.Name >= name
		} else if entry.DirEntries != nil {
			return entry.DirEntries.MinName() >= name
		}
		panic("impossible")
	})

	defer func() {
		if newEntries != nil {
			//TODO compact
		}
	}()

	if i >= len(d) {
		// not found
		file, err := (*File)(nil).Apply(path[1:], op)
		if err != nil {
			return nil, err
		}
		if file != nil {
			// append
			newEntries := make(DirEntries, len(d))
			copy(newEntries, d)
			newEntries = append(newEntries, DirEntry{
				File: file,
			})
			return &newEntries, nil
		}
		return nil, nil
	}

	entry := d[i]
	if entry.File != nil {
		if entry.File.Name != name {
			// not found
			file, err := (*File)(nil).Apply(path[1:], op)
			if err != nil {
				return nil, err
			}
			if file != nil {
				// insert
				newEntries := make(DirEntries, 0, len(d)+1)
				newEntries = append(newEntries, d[:i]...)
				newEntries = append(newEntries, DirEntry{
					File: file,
				})
				newEntries = append(newEntries, d[i:]...)
				return &newEntries, nil
			}
			return nil, nil

		} else {
			// found
			newFile, err := entry.File.Apply(path[1:], op)
			if err != nil {
				return nil, err
			}
			if err := entry.File.checkNewFile(newFile); err != nil {
				return nil, err
			}
			if newFile == nil {
				// delete
				newEntries := make(DirEntries, 0, len(d)-1)
				newEntries = append(newEntries, d[:i]...)
				newEntries = append(newEntries, d[i+1:]...)
				return &newEntries, nil
			} else if newFile != entry.File {
				// replace
				newEntries := make(DirEntries, 0, len(d))
				newEntries = append(newEntries, d[:i]...)
				newEntries = append(newEntries, DirEntry{
					File: newFile,
				})
				newEntries = append(newEntries, d[:i+1]...)
				return &newEntries, nil
			}
			return nil, nil
		}

	} else if entry.DirEntries != nil {
		if entry.DirEntries.MinName() > name {
			// not found
			file, err := (*File)(nil).Apply(path[1:], op)
			if err != nil {
				return nil, err
			}
			if file != nil {
				// insert
				newEntries := make(DirEntries, 0, len(d)+1)
				newEntries = append(newEntries, d[:i]...)
				newEntries = append(newEntries, DirEntry{
					File: file,
				})
				newEntries = append(newEntries, d[i:]...)
				return &newEntries, nil
			}
			return nil, nil
		}

		newSubEntries, err := entry.DirEntries.Apply(path, op)
		if err != nil {
			return nil, err
		}
		if newSubEntries != nil {
			// replace
			newEntries := make(DirEntries, 0, len(d))
			newEntries = append(newEntries, d[:i]...)
			newEntries = append(newEntries, DirEntry{
				DirEntries: newSubEntries,
			})
			newEntries = append(newEntries, d[i+1:]...)
			return &newEntries, nil
		}
		return nil, nil
	}

	panic("impossible")
}
