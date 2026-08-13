package huffman

type Code struct {
	Bits  uint8  // The number of bits that are used in the Value field to represent the code
	Value uint32 // Value is the representation of the code in bits, stored in the least significant bits
}

// WithAppendedBit returns a new Code with the given bit appended to the end of the current code.
func (c Code) WithAppendedBit(bit bool) Code {
	cc := Code{
		Bits:  c.Bits + 1,
		Value: c.Value << 1,
	}
	if bit {
		cc.Value |= 1
	}
	return cc
}
