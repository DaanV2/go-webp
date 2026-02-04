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

const VP8L_LOG8_WBITS = 4 // Number of bytes needed to store VP8L_WBITS bits.

// This is the minimum amount of size the memory buffer is guaranteed to grow
// when extra space is needed.
const MIN_EXTRA_SIZE =(uint64(32768))



const VP8L_WRITER_BYTES =4      // sizeof(vp8l_wtype_t)
const VP8L_WRITER_BITS =32      // 8 * sizeof(vp8l_wtype_t)
const VP8L_WRITER_MAX_BITS =64  // 8 * sizeof(vp8l_atype_t)

var kVP8Log2Range = [128]uint8{
	7, 6, 6, 5, 5, 5, 5, 4, 4, 4, 4, 4, 4, 4, 4, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0}

// range = ((range - 1) << kVP8Log2Range[range]) + 1
var kVP8NewRange = [128]uint8{
	127, 127, 191, 127, 159, 191, 223, 127, 143, 159, 175, 191, 207, 223, 239, 127, 135, 143, 151, 159, 167, 175, 183, 191, 199, 207, 215, 223, 231, 239, 247, 127, 131, 135, 139, 143, 147, 151, 155, 159, 163, 167, 171, 175, 179, 183, 187, 191, 195, 199, 203, 207, 211, 215, 219, 223, 227, 231, 235, 239, 243, 247, 251, 127, 129, 131, 133, 135, 137, 139, 141, 143, 145, 147, 149, 151, 153, 155, 157, 159, 161, 163, 165, 167, 169, 171, 173, 175, 177, 179, 181, 183, 185, 187, 189, 191, 193, 195, 197, 199, 201, 203, 205, 207, 209, 211, 213, 215, 217, 219, 221, 223, 225, 227, 229, 231, 233, 235, 237, 239, 241, 243, 245, 247, 249, 251, 253, 127}

var kBitMask = [VP8L_MAX_NUM_BIT_READ + 1]uint32{
	0, 0x000001, 0x000003, 0x000007, 0x00000f, 0x00001f, 0x00003f, 0x00007f, 0x0000ff, 0x0001ff, 0x0003ff, 0x0007ff, 0x000fff, 0x001fff, 0x003fff, 0x007fff, 0x00ffff, 0x01ffff, 0x03ffff, 0x07ffff, 0x0fffff, 0x1fffff, 0x3fffff, 0x7fffff, 0xffffff}

var kNorm = [128]uint8{ // renorm_sizes[i] = 8 - log2(i)
	7, 6, 6, 5, 5, 5, 5, 4, 4, 4, 4, 4, 4, 4, 4, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0}

// range = ((range + 1) << kVP8Log2Range[range]) - 1
var kNewRange = [128]uint8{
	127, 127, 191, 127, 159, 191, 223, 127, 143, 159, 175, 191, 207, 223, 239, 127, 135, 143, 151, 159, 167, 175, 183, 191, 199, 207, 215, 223, 231, 239, 247, 127, 131, 135, 139, 143, 147, 151, 155, 159, 163, 167, 171, 175, 179, 183, 187, 191, 195, 199, 203, 207, 211, 215, 219, 223, 227, 231, 235, 239, 243, 247, 251, 127, 129, 131, 133, 135, 137, 139, 141, 143, 145, 147, 149, 151, 153, 155, 157, 159, 161, 163, 165, 167, 169, 171, 173, 175, 177, 179, 181, 183, 185, 187, 189, 191, 193, 195, 197, 199, 201, 203, 205, 207, 209, 211, 213, 215, 217, 219, 221, 223, 225, 227, 229, 231, 233, 235, 237, 239, 241, 243, 245, 247, 249, 251, 253, 127}

