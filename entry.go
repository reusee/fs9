package fs9

type Entry struct {
	*File
	*Entries
}

type Entries []Entry
