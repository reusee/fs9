package fs9

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/reusee/e4"
	"github.com/reusee/pp"
)

// DirEntry is a versioned fat-node
type DirEntry []DirEntryValue

type DirEntryValue struct {
	version    int64
	File       *File
	DirEntries *DirEntries
}

const maxDirEntryLen = 4

func (d DirEntry) Latest() any {
	value := d[len(d)-1]
	if value.File != nil {
		return value.File
	} else if value.DirEntries != nil {
		return value.DirEntries
	}
	panic("impossible")
}

type DirEntries []DirEntry

func (d DirEntries) MinName() string {
	if len(d) == 0 {
		panic("impossible")
	}
	switch item := d[0].Latest().(type) {
	case *File:
		return item.Name
	case *DirEntries:
		return item.MinName()
	}
	panic("impossible")
}

func (d *DirEntries) Apply(path []string, ctx OperationCtx, op Operation) (newEntries *DirEntries, err error) {
	//ce(d.verifyStructure())
	//defer func() {
	//	if newEntries != nil {
	//		ce(newEntries.verifyStructure())
	//	}
	//}()

	we := we.With(
		e4.Info("path: %s", strings.Join(path, "/")),
	)

	if len(path) == 0 {
		return nil, we(ErrFileNotFound)
	}

	// name
	name := path[0]
	if name == "" || name == "." || name == ".." {
		return nil, we(ErrInvalidName)
	}

	entries := *d

	// descend
	i := sort.Search(len(entries), func(i int) bool {
		switch item := entries[i].Latest().(type) {
		case *File:
			return item.Name >= name
		case *DirEntries:
			return item.MinName() >= name
		}
		panic("impossible")
	})

	defer func() {
		if newEntries != nil {
			//TODO compact
		}
	}()

	if i == len(entries) {
		// not found
		file, err := (*File)(nil).Apply(path[1:], ctx, op)
		if err != nil {
			return nil, err
		}
		if file != nil {
			// append
			newEntries := make(DirEntries, len(entries), len(entries)+1)
			copy(newEntries, entries)
			newEntries = append(newEntries, DirEntry{
				{
					version: ctx.Version,
					File:    file,
				},
			})
			return &newEntries, nil
		}
		return d, nil
	}

	switch item := entries[i].Latest().(type) {

	case *File:
		if item.Name != name {
			// not found
			file, err := (*File)(nil).Apply(path[1:], ctx, op)
			if err != nil {
				return nil, err
			}
			if file != nil {
				// insert
				newEntries := make(DirEntries, 0, len(entries)+1)
				newEntries = append(newEntries, entries[:i]...)
				newEntries = append(newEntries, DirEntry{
					{
						version: ctx.Version,
						File:    file,
					},
				})
				newEntries = append(newEntries, entries[i:]...)
				return &newEntries, nil
			}
			return d, nil

		} else {
			// found
			newFile, err := item.Apply(path[1:], ctx, op)
			if err != nil {
				return nil, err
			}
			if err := item.checkNewFile(newFile); err != nil {
				return nil, err
			}
			if newFile == nil {
				// delete
				newEntries := make(DirEntries, 0, len(entries)-1)
				newEntries = append(newEntries, entries[:i]...)
				newEntries = append(newEntries, entries[i+1:]...)
				return &newEntries, nil
			} else if newFile != item {
				// replace
				if len(entries[i]) < maxDirEntryLen {
					entries[i] = append(entries[i], DirEntryValue{
						version: ctx.Version,
						File:    newFile,
					})
					return d, nil
				} else {
					newEntries := make(DirEntries, 0, len(entries))
					newEntries = append(newEntries, entries[:i]...)
					newEntries = append(newEntries, DirEntry{
						{
							version: ctx.Version,
							File:    newFile,
						},
					})
					newEntries = append(newEntries, entries[i+1:]...)
					return &newEntries, nil
				}
			}
			return d, nil
		}

	case *DirEntries:
		if item.MinName() > name {
			// not found
			file, err := (*File)(nil).Apply(path[1:], ctx, op)
			if err != nil {
				return nil, err
			}
			if file != nil {
				// insert
				newEntries := make(DirEntries, 0, len(entries)+1)
				newEntries = append(newEntries, entries[:i]...)
				newEntries = append(newEntries, DirEntry{
					{
						version: ctx.Version,
						File:    file,
					},
				})
				newEntries = append(newEntries, entries[i:]...)
				return &newEntries, nil
			}
			return d, nil
		}

		newSubEntries, err := item.Apply(path, ctx, op)
		if err != nil {
			return nil, err
		}
		if newSubEntries == nil {
			// remove
			newEntries := make(DirEntries, 0, len(entries)-1)
			newEntries = append(newEntries, entries[:i]...)
			newEntries = append(newEntries, entries[i+1:]...)
			return &newEntries, nil
		} else if newSubEntries != item {
			// replace
			if len(entries[i]) < maxDirEntryLen {
				entries[i] = append(entries[i], DirEntryValue{
					version:    ctx.Version,
					DirEntries: newSubEntries,
				})
				return d, nil
			} else {
				newEntries := make(DirEntries, 0, len(entries))
				newEntries = append(newEntries, entries[:i]...)
				newEntries = append(newEntries, DirEntry{
					{
						version:    ctx.Version,
						DirEntries: newSubEntries,
					},
				})
				newEntries = append(newEntries, entries[i+1:]...)
				return &newEntries, nil
			}
		}
		return d, nil

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
	return e4.Info("%s", buf.String())
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
		switch a := a.Latest().(type) {
		case *File:
			switch b := b.Latest().(type) {
			case *File:
				return a.Name < b.Name
			case *DirEntries:
				return a.Name < b.MinName()
			}
			panic("impossible")
		case *DirEntries:
			switch b := b.Latest().(type) {
			case *File:
				return a.MinName() < b.Name
			case *DirEntries:
				return a.MinName() < b.MinName()
			}
		}
		panic("impossible")
	})
	for i, j := range idx {
		if i != j {
			return ce.With(
				d.wrapDump(),
				e4.Info("order: %+v", idx),
			)(fmt.Errorf("invalid order"))
		}
	}

	return nil
}
