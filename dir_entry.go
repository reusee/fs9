package fs9

type DirEntry struct {
	*File
	*DirEntries
}

type DirEntries []DirEntry
