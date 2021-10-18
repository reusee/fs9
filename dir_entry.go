package fs9

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/reusee/e4"
	"github.com/reusee/pp"
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
	//TODO
	ce(d.verifyStructure())
	defer func() {
		if newEntries != nil {
			ce(newEntries.verifyStructure())
		}
	}()

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

	if i == len(d) {
		// not found
		file, err := (*File)(nil).Apply(path[1:], op)
		if err != nil {
			return nil, err
		}
		if file != nil {
			// append
			newEntries := make(DirEntries, len(d), len(d)+1)
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
				newEntries = append(newEntries, d[i+1:]...)
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

func (d DirEntries) Dump(w io.Writer, level int) {
	ce(pp.Copy(
		d.IterFiles(nil),
		pp.Tap(func(v any) error {
			v.(*File).Dump(w, level)
			return nil
		}),
	))
}

func (d DirEntries) wrapDump() e4.WrapFunc {
	buf := new(strings.Builder)
	d.Dump(buf, 0)
	return e4.NewInfo("%s", buf.String())
}

func (d DirEntries) verifyStructure() error {

	// names
	names := make(map[string]bool)
	ce(pp.Copy(
		d.IterFiles(nil),
		pp.Tap(func(v any) error {
			name := v.(*File).Name
			if _, ok := names[name]; ok {
				return we.With(
					d.wrapDump(),
				)(fmt.Errorf("duplicated name"))
			}
			names[name] = true
			return nil
		}),
	))

	// order
	var idx []int
	for i := range d {
		idx = append(idx, i)
	}
	sort.SliceStable(idx, func(i, j int) bool {
		a := d[i]
		b := d[j]
		if a.File != nil {
			if b.File != nil {
				return a.File.Name < b.File.Name
			} else if b.DirEntries != nil {
				return a.File.Name < b.DirEntries.MinName()
			}
			panic("impossible")
		} else if a.DirEntries != nil {
			if b.File != nil {
				return a.DirEntries.MinName() < b.File.Name
			} else if b.DirEntries != nil {
				return a.DirEntries.MinName() < b.DirEntries.MinName()
			}
		}
		panic("impossible")
	})
	for i, j := range idx {
		if i != j {
			return ce.With(
				d.wrapDump(),
				e4.NewInfo("order: %+v", idx),
			)(fmt.Errorf("invalid order"))
		}
	}

	return nil
}
