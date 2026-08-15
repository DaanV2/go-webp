package riff

import "bytes"

// FourCC or Four Character Code is the id of an chunk
type FourCC [4]byte

func (f FourCC) Len() int {
	return len(f)
}

func (f FourCC) EqualString(other string) bool {
	return bytes.Equal(f[:], []byte(other))
}

func (f FourCC) Equal(other FourCC) bool {
	return bytes.Equal(f[:], other[:])
}

func (f FourCC) String() string {
	return string(f[:])
}
