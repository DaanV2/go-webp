package bits

type lbit_t = uint64

// #if (BITS > 24)
// ------------------------------------------------------------------------------
// Derived types and constants:
//
//	bit_t = natural register type for storing 'value' (which is BITS+8 bits)
//	range_t = register for 'range' (which is 8bits only)
type bit_t uint64

// #else
// typedef uint32 bit_t;
// #endif

type range_t uint32
type vp8l_val_t uint64   // right now, this bit-reader can only use 64bit.
type vp8l_atype_t uint64 // accumulator type
type vp8l_wtype_t uint32 // writing type
