package constants

import (
	"encoding/binary"
)

var WORDS_BIGENDIAN bool

func init() {
	// Determine endianness at runtime
	const v = 0x0102030405060708

	var testb, bigb []byte = make([]byte, 8), make([]byte, 8)
	binary.NativeEndian.PutUint64(testb, v)
	binary.BigEndian.PutUint64(bigb, v)

	WORDS_BIGENDIAN = string(testb) == string(bigb)
}
