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

// typedef uint32 range_t;
