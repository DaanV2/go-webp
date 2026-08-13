package huffman

type BitWriter interface {
	// WriteBit writes a single bit to the output stream. The bit is represented as a boolean value, where true represents 1 and false represents 0.
	WriteBit(bit bool) error
}

type BitsWriter interface {
	// WriteBits writes the specified number of bits from the given value to the output stream. The bits are taken from the least significant bits of the value.
	WriteBits(bits uint32, count uint8) error
}

type BitReader interface {
	// ReadBit reads a single bit from the input stream and returns it as a boolean value, where true represents 1 and false represents 0.
	ReadBit() (bool, error)
}

type BitsReader interface {
	// ReadBits reads the specified number of bits from the input stream and returns them as a uint32 value. The bits are read from the least significant bits of the returned value.
	ReadBits(count uint8) (uint32, error)
}
