package fs9

func (d DirEntries) IterFileInfos(cont Src) Src {
	return func() (any, Src, error) {
		if len(d) == 0 {
			return nil, cont, nil
		}
		switch item := d[0].Latest().(type) {
		case *File:
			return item.Info(), d[1:].IterFileInfos(cont), nil
		case *DirEntries:
			return nil, item.IterFileInfos(
				d[:1].IterFileInfos(cont),
			), nil
		}
		panic("impossible")
	}
}

func (d DirEntries) IterFiles(cont Src) Src {
	return func() (any, Src, error) {
		if len(d) == 0 {
			return nil, cont, nil
		}
		switch item := d[0].Latest().(type) {
		case *File:
			return item, d[1:].IterFiles(cont), nil
		case *DirEntries:
			return nil, item.IterFiles(
				d[:1].IterFiles(cont),
			), nil
		}
		panic("impossible")
	}
}
