package fs9

import (
	"sort"

	"github.com/reusee/e4"
)

type Mutator func(
	ctx MutateCtx,
	entry *DirEntry,
) (
	newEntry *DirEntry,
	err error,
)

type MutateCtx struct {
	Version int64
}

func MutatePath(path []string, fn Mutator) Mutator {
	return func(ctx MutateCtx, entry *DirEntry) (*DirEntry, error) {
		we := we.With(
			e4.Info("path: %+v", path),
		)

		if len(path) == 0 {
			// reach target
			newEntry, err := fn(ctx, entry)
			if err != nil {
				return nil, we(err)
			}
			return newEntry, nil
		}

		// descend
		ptr, ok := entry.Latest().(*DirEntries)
		if !ok {
			return nil, we(ErrFileNotFound)
		}
		entries := *ptr

		name := path[0]
		if name == "" || name == "." || name == ".." {
			return nil, we(ErrInvalidName)
		}

		idx := sort.Search(len(entries), func(i int) bool {
			switch item := entries[i].Latest().(type) {
			case *File:
				return item.Name >= name
			case *DirEntries:
				return item.MinName() >= name
			}
			panic("impossible")
		})

		if idx == len(entries) {
			// not found
			sub, err := MutatePath(path[1:], fn)(ctx, nil)
			if err != nil {
				return nil, we(err)
			}
			if sub != nil {
				// append
				newEntries := make(DirEntries, len(entries), len(entries)+1)
				copy(newEntries, entries)
				newEntries = append(newEntries, *sub)
				return &DirEntry{
					{
						version:    ctx.Version,
						DirEntries: &newEntries,
					},
				}, nil
			}
			// not changed
			return entry, nil
		}

		switch item := entries[idx].Latest().(type) {

		case *File:
			if item.Name != name {
				// not found
				sub, err := MutatePath(path[1:], fn)(ctx, nil)
				if err != nil {
					return nil, we(err)
				}
				if sub != nil {
					// insert
					newEntries := make(DirEntries, 0, len(entries)+1)
					newEntries = append(newEntries, entries[:idx]...)
					newEntries = append(newEntries, *sub)
					newEntries = append(newEntries, entries[idx:]...)
					return &DirEntry{
						{
							version:    ctx.Version,
							DirEntries: &newEntries,
						},
					}, nil
				}
				// not changed
				return entry, nil
			}

			// found
			sub, err := MutatePath(path[1:], fn)(ctx, &entries[idx])
			if err != nil {
				return nil, we(err)
			}
			if sub == nil {
				// delete
				newEntries := make(DirEntries, 0, len(entries)-1)
				newEntries = append(newEntries, entries[:idx]...)
				newEntries = append(newEntries, entries[idx+1:]...)
				return &DirEntry{
					{
						version:    ctx.Version,
						DirEntries: &newEntries,
					},
				}, nil
			} else if sub != &entries[idx] {
				// replace
				if len(entries[idx]) < maxDirEntryLen {
					// add to fat-node
					entries[idx] = append(entries[idx], DirEntryValue{
						version: ctx.Version,
						File:    sub.Latest().(*File),
					})
					return entry, nil
				} else {
					// path copy
					newEntries := make(DirEntries, 0, len(entries))
					newEntries = append(newEntries, entries[:idx]...)
					newEntries = append(newEntries, *sub)
					newEntries = append(newEntries, entries[idx+1:]...)
					return &DirEntry{
						{
							version:    ctx.Version,
							DirEntries: &newEntries,
						},
					}, nil
				}
			}
			// no change
			return entry, nil

		case *DirEntries:
			if item.MinName() > name {
				// not found
				sub, err := MutatePath(path[1:], fn)(ctx, nil)
				if err != nil {
					return nil, we(err)
				}
				if sub != nil {
					// insert
					newEntries := make(DirEntries, 0, len(entries)+1)
					newEntries = append(newEntries, entries[:idx]...)
					newEntries = append(newEntries, *sub)
					newEntries = append(newEntries, entries[idx:]...)
					return &DirEntry{
						{
							version:    ctx.Version,
							DirEntries: &newEntries,
						},
					}, nil
				}
				// no change
				return entry, nil
			}

			// found
			sub, err := MutatePath(path[1:], fn)(ctx, &entries[idx])
			if err != nil {
				return nil, we(err)
			}
			if sub == nil {
				// remove
				newEntries := make(DirEntries, 0, len(entries)-1)
				newEntries = append(newEntries, entries[:idx]...)
				newEntries = append(newEntries, entries[idx+1:]...)
				return &DirEntry{
					{
						version:    ctx.Version,
						DirEntries: &newEntries,
					},
				}, nil
			} else if sub != &entries[idx] {
				// replace
				if len(entries[idx]) < maxDirEntryLen {
					// append to fat-node
					entries[idx] = append(entries[idx], DirEntryValue{
						version:    ctx.Version,
						DirEntries: sub.Latest().(*DirEntries),
					})
					return entry, nil
				} else {
					// copy path
					newEntries := make(DirEntries, 0, len(entries))
					newEntries = append(newEntries, entries[:idx]...)
					newEntries = append(newEntries, *sub)
					newEntries = append(newEntries, entries[idx+1:]...)
					return &DirEntry{
						{
							version:    ctx.Version,
							DirEntries: &newEntries,
						},
					}, nil
				}
			}

			// no change
			return entry, nil
		}

		panic("impossible")
	}
}
