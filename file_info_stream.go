package fs9

func (d DirEntries) IterFileInfos(cont Src) Src {
	return func() (any, Src, error) {
		if len(d) == 0 {
			return nil, cont, nil
		}
		item := d[0]
		if item.File != nil {
			return item.File.Info(), d[1:].IterFileInfos(cont), nil
		} else if item.DirEntries != nil {
			return nil, item.DirEntries.IterFileInfos(
				d[:1].IterFileInfos(cont),
			), nil
		}
		panic("impossible")
	}
}
