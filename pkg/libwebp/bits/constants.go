package bits

// The Boolean decoder needs to maintain infinite precision on the 'value'
// field. However, since 'range' is only 8bit, we only need an active window of
// 8 bits for 'value". Left bits (MSB) gets zeroed and shifted away when
// 'value' falls below 128, 'range' is updated, and fresh bits read from the
// bitstream are brought in as LSB. To afunc reading the fresh bits one by one
// (slow), we cache BITS of them ahead. The total of (BITS + 8) bits must fit
// into a natural register (with type bit_t). To fetch BITS bits from bitstream
// we use a type lbit_t.
//
// BITS can be any multiple of 8 from 8 to 56 (inclusive).
// Pick values that fit natural register size.

// #if defined(__i386__) || defined(_M_IX86)  // x86 32bit
// const BITS = 24
// #elif defined(__x86_64__) || defined(_M_X64)  // x86 64bit
// const BITS = 56
// #elif defined(__arm__) || defined(_M_ARM)  // ARM
// const BITS = 24
// #elif WEBP_AARCH64  // ARM 64bit
// const BITS = 56
// #elif defined(__mips__)  // MIPS
// const BITS = 24
// #elif defined(__wasm__)  // WASM
// const BITS = 56
// #else  // reasonable default

// #endif

const BITS = 56
const VP8L_MAX_NUM_BIT_READ = 24

const VP8L_LBITS = 64 // Number of bits prefetched (= bit-size of vp8l_val_t).
const VP8L_WBITS = 32 // Minimum number of bytes ready after VP8LFillBitWindow.

// #if (BITS > 24)
//------------------------------------------------------------------------------
// Derived types and constants:
//   bit_t = natural register type for storing 'value' (which is BITS+8 bits)
//   range_t = register for 'range' (which is 8bits only)
type bit_t uint64

// #else
// typedef uint32 bit_t;
// #endif

// typedef uint32 range_t;
