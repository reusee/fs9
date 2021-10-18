package fs9

import (
	"crypto/rand"
	"encoding/binary"
	mathrand "math/rand"
)

func init() {
	var seed int64
	ce(binary.Read(rand.Reader, binary.LittleEndian, &seed))
	mathrand.Seed(seed)
}
