package vp8

// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// main entry for the decoder
//
// Authors: Vikas Arora (vikaas.arora@gmail.com)
//          Jyrki Alakuijala (jyrki@google.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stddef"
import "github.com/daanv2/go-webp/pkg/stdlib"
// import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/decoder"
import "github.com/daanv2/go-webp/pkg/libwebp/decoder"
import "github.com/daanv2/go-webp/pkg/libwebp/decoder"
import "github.com/daanv2/go-webp/pkg/libwebp/decoder"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


const NUM_ARGB_CACHE_ROWS =16

const kCodeLengthLiterals = 16
const kCodeLengthRepeatCode = 16
var kCodeLengthExtraBits  = [3]uint8{2, 3, 7}
var kCodeLengthRepeatOffsets  = [3]uint8{3, 3, 11}

// -----------------------------------------------------------------------------
//  Five Huffman codes are used at each meta code:
//  1. green + length prefix codes + color cache codes,
//  2. alpha,
//  3. red,
//  4. blue, and,
//  5. distance prefix codes.
type HuffIndex int

const ( 
	GREEN HuffIndex = 0
	RED HuffIndex = 1
	BLUE HuffIndex = 2
	ALPHA HuffIndex = 3
	DIST HuffIndex = 4
)

var kAlphabetSize = [constants.HUFFMAN_CODES_PER_META_CODE]uint16{
    NUM_LITERAL_CODES + NUM_LENGTH_CODES, NUM_LITERAL_CODES, NUM_LITERAL_CODES, NUM_LITERAL_CODES, NUM_DISTANCE_CODES,
}

var kLiteralMap = [constants.HUFFMAN_CODES_PER_META_CODE]uint8{0, 1, 1, 1, 0}

const NUM_CODE_LENGTH_CODES = 19
const CODE_TO_PLANE_CODES =120

var kCodeLengthCodeOrder = [NUM_CODE_LENGTH_CODES]uint8{
    17, 18, 0, 1, 2, 3, 4, 5, 16, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

var kCodeToPlane = [CODE_TO_PLANE_CODES]uint8{
    0x18, 0x07, 0x17, 0x19, 0x28, 0x06, 0x27, 0x29, 0x16, 0x1a, 0x26, 0x2a, 0x38, 0x05, 0x37, 0x39, 0x15, 0x1b, 0x36, 0x3a, 0x25, 0x2b, 0x48, 0x04, 0x47, 0x49, 0x14, 0x1c, 0x35, 0x3b, 0x46, 0x4a, 0x24, 0x2c, 0x58, 0x45, 0x4b, 0x34, 0x3c, 0x03, 0x57, 0x59, 0x13, 0x1d, 0x56, 0x5a, 0x23, 0x2d, 0x44, 0x4c, 0x55, 0x5b, 0x33, 0x3d, 0x68, 0x02, 0x67, 0x69, 0x12, 0x1e, 0x66, 0x6a, 0x22, 0x2e, 0x54, 0x5c, 0x43, 0x4d, 0x65, 0x6b, 0x32, 0x3e, 0x78, 0x01, 0x77, 0x79, 0x53, 0x5d, 0x11, 0x1f, 0x64, 0x6c, 0x42, 0x4e, 0x76, 0x7a, 0x21, 0x2f, 0x75, 0x7b, 0x31, 0x3f, 0x63, 0x6d, 0x52, 0x5e, 0x00, 0x74, 0x7c, 0x41, 0x4f, 0x10, 0x20, 0x62, 0x6e, 0x30, 0x73, 0x7d, 0x51, 0x5f, 0x40, 0x72, 0x7e, 0x61, 0x6f, 0x50, 0x71, 0x7f, 0x60, 0x70}

// Memory needed for lookup tables of one Huffman tree group. Red, blue, alpha
// and distance alphabets are constant (256 for red, blue and alpha, 40 for
// distance) and lookup table sizes for them in worst case are 630 and 410
// respectively. Size of green alphabet depends on color cache size and is equal
// to 256 (green component values) + 24 (length prefix values)
// + color_cache_size (between 0 and 2048).
// All values computed for 8-bit first level lookup with Mark Adler's tool:
// https://github.com/madler/zlib/blob/v1.2.5/examples/enough.c
const FIXED_TABLE_SIZE =(630 * 3 + 410)
var  kTableSize = [12]uint16{
    FIXED_TABLE_SIZE + 654,  FIXED_TABLE_SIZE + 656,  FIXED_TABLE_SIZE + 658, FIXED_TABLE_SIZE + 662,  FIXED_TABLE_SIZE + 670,  FIXED_TABLE_SIZE + 686, FIXED_TABLE_SIZE + 718,  FIXED_TABLE_SIZE + 782,  FIXED_TABLE_SIZE + 912, FIXED_TABLE_SIZE + 1168, FIXED_TABLE_SIZE + 1680, FIXED_TABLE_SIZE + 2704}

func VP8LSetError(/* const */ dec *VP8LDecoder, error VP8StatusCode) int {
  // The oldest error reported takes precedence over the new one.
  if dec.status == VP8_STATUS_OK || dec.status == VP8_STATUS_SUSPENDED {
    dec.status = error
  }
  return 0
}

func DecodeImageStream(xsize int , ysize int , is_level0 int, /* const */ dec *VP8LDecoder, decoded_data *uint32) int {
	// TODO: implementation
	return 0
}

//------------------------------------------------------------------------------

// Returns true if the next byte(s) in data is a VP8L signature.
func VP8LCheckSignature(/* const */  data *uint8 , size uint64) int {
  return (size >= VP8L_FRAME_HEADER_SIZE && data[0] == VP8L_MAGIC_BYTE &&
          (data[4] >> 5) == 0);  // version
}

func ReadImageInfo(/* const */ br *VP8LBitReader, /* const */ width *int, /* const */ height *int, /* const */ has_alpha *int) int {
  if VP8LReadBits(br, 8) != VP8L_MAGIC_BYTE {return 0}
  *width = VP8LReadBits(br, VP8L_IMAGE_SIZE_BITS) + 1
  *height = VP8LReadBits(br, VP8L_IMAGE_SIZE_BITS) + 1
  *has_alpha = VP8LReadBits(br, 1)
  if VP8LReadBits(br, VP8L_VERSION_BITS) != 0 {return 0}
  return !br.eos
}

// Validates the VP8L data-header and retrieves basic header information viz
// width, height and alpha. Returns 0 in case of formatting error.
// width/height/has_alpha can be passed nil.
func VP8LGetInfo(/* const */ data *uint8, data_size uint64 , /* const */ width *int, /* const */ height *int, /* const */ has_alpha *int) int {
  if data == nil || data_size < VP8L_FRAME_HEADER_SIZE {
    return 0;  // not enough data
  } else if !VP8LCheckSignature(data, data_size) {
    return 0;  // bad signature
  } else {
    var w, h, a int
    var br VP8LBitReader
    VP8LInitBitReader(&br, data, data_size)
    if !ReadImageInfo(&br, &w, &h, &a) {
      return 0
    }
    if width != nil { {*width = w }}
    if height != nil { {*height = h }}
    if has_alpha != nil { {*has_alpha = a }}
    return 1
  }
}

func GetCopyDistance(distance_symbol int, /* const */ br *VP8LBitReader) int {
  var extra_bits, offset int
  if distance_symbol < 4 {
    return distance_symbol + 1
  }
  extra_bits = (distance_symbol - 2) >> 1
  offset = (2 + (distance_symbol & 1)) << extra_bits
  return offset + VP8LReadBits(br, extra_bits) + 1
}

func GetCopyLength( length_symbol int, br *VP8LBitReader) int {
  // Length and distance prefixes are encoded the same way.
  return GetCopyDistance(length_symbol, br)
}

func PlaneCodeToDistance(xsize, plane_code int ) int {
  if plane_code > CODE_TO_PLANE_CODES {
    return plane_code - CODE_TO_PLANE_CODES
  } else {
    dist_code := kCodeToPlane[plane_code - 1]
    yoffset := dist_code >> 4
    xoffset := 8 - (dist_code & 0xf)
    dist := yoffset * xsize + xoffset
    return tenary.If(dist >= 1, dist, 1);  // dist<1 can happen if xsize is very small
  }
}

//------------------------------------------------------------------------------
// Decodes the next Huffman code from bit-stream.
// VP8LFillBitWindow(br) needs to be called at minimum every second call
// to ReadSymbol, in order to pre-fetch enough bits.
func ReadSymbol(/* const */ table *HuffmanCode, /*const*/ br *VP8LBitReader) int {
  var nbits int
  val := VP8LPrefetchBits(br)
  table += val & HUFFMAN_TABLE_MASK
  nbits = table.bits - HUFFMAN_TABLE_BITS
  if nbits > 0 {
    VP8LSetBitPos(br, br.bit_pos + HUFFMAN_TABLE_BITS)
    val = VP8LPrefetchBits(br)
    table += table.value
    table += val & ((1 << nbits) - 1)
  }
  VP8LSetBitPos(br, br.bit_pos + table.bits)
  return table.value
}

// Reads packed symbol depending on GREEN channel
const BITS_SPECIAL_MARKER =0x100  // something large enough (and a bit-mask)
const PACKED_NON_LITERAL_CODE =0  // must be < NUM_LITERAL_CODES
func ReadPackedSymbols(/* const */ group *HTreeGroup, /*const*/ br *VP8LBitReader, /*const*/ dst *uint32) int {
  val := VP8LPrefetchBits(br) & (HUFFMAN_PACKED_TABLE_SIZE - 1)
  var code HuffmanCode32 = group.packed_table[val]
  assert.Assert(group.use_packed_table)
  if code.bits < BITS_SPECIAL_MARKER {
    VP8LSetBitPos(br, br.bit_pos + code.bits)
    *dst = code.value
    return PACKED_NON_LITERAL_CODE
  } else {
    VP8LSetBitPos(br, br.bit_pos + code.bits - BITS_SPECIAL_MARKER)
    assert.Assert(code.value >= NUM_LITERAL_CODES)
    return code.value
  }
}

func AccumulateHCode(HuffmanCode hcode, shift int, /*const*/ huff *HuffmanCode32) int {
  huff.bits += hcode.bits
  huff.value |= uint32(hcode.value) << shift
  assert.Assert(huff.bits <= HUFFMAN_TABLE_BITS)
  return hcode.bits
}

func BuildPackedTable(/* const */ htree_group *HTreeGroup) {
  var code uint32
  for code = 0; code < HUFFMAN_PACKED_TABLE_SIZE; code++ {
    bits := code
    var huff *HuffmanCode32 = &htree_group.packed_table[bits]
    var hcode HuffmanCode = htree_group.htrees[GREEN][bits]
    if hcode.value >= NUM_LITERAL_CODES {
      huff.bits = hcode.bits + BITS_SPECIAL_MARKER
      huff.value = hcode.value
    } else {
      huff.bits = 0
      huff.value = 0
      bits >>= AccumulateHCode(hcode, 8, huff)
      bits >>= AccumulateHCode(htree_group.htrees[RED][bits], 16, huff)
      bits >>= AccumulateHCode(htree_group.htrees[BLUE][bits], 0, huff)
      bits >>= AccumulateHCode(htree_group.htrees[ALPHA][bits], 24, huff)
      // C: (void)bits
    }
  }
}

func ReadHuffmanCodeLengths(/* const */ dec *VP8LDecoder, /*const*/ code_length_code_lengths *int, num_symbols int, /*const*/ code_lengths *int) int {
  ok := 0
  var br *VP8LBitReader = &decoder.br
  var symbol int
  var max_symbol int
  prev_code_len := DEFAULT_CODE_LENGTH
   var tables HuffmanTables
  var bounded_code_lengths *int = WEBP_UNSAFE_FORGE_BIDI_INDEXABLE(
          // C: const *int, code_length_code_lengths, NUM_CODE_LENGTH_CODES * sizeof(*code_length_code_lengths))

  if (!VP8LHuffmanTablesAllocate(1 << LENGTHS_TABLE_BITS, &tables) ||
      !VP8LBuildHuffmanTable(&tables, LENGTHS_TABLE_BITS, bounded_code_lengths, NUM_CODE_LENGTH_CODES)) {
    goto End
  }

  if VP8LReadBits(br, 1) {  // use length
    length_nbits := 2 + 2 * VP8LReadBits(br, 3)
    max_symbol = 2 + VP8LReadBits(br, length_nbits)
    if max_symbol > num_symbols {
      goto End
    }
  } else {
    max_symbol = num_symbols
  }

  symbol = 0
  for symbol < num_symbols {
    const p *HuffmanCode
    var code_len int
    if max_symbol-- == 0 { break }
    VP8LFillBitWindow(br)
    p = &tables.curr_segment.start[VP8LPrefetchBits(br) & LENGTHS_TABLE_MASK]
    VP8LSetBitPos(br, br.bit_pos + p.bits)
    code_len = p.value
    if code_len < kCodeLengthLiterals {
		symbol++
      code_lengths[symbol] = code_len
      if code_len != 0 { prev_code_len = code_len }
    } else {
      use_prev := (code_len == kCodeLengthRepeatCode)
      slot := code_len - kCodeLengthLiterals
      extra_bits := kCodeLengthExtraBits[slot]
      repeat_offset := kCodeLengthRepeatOffsets[slot]
      repeat := VP8LReadBits(br, extra_bits) + repeat_offset
      if symbol + repeat > num_symbols {
        goto End
      } else {
        length := tenary.If(use_prev, prev_code_len,  0)
        for repeat > 0 {
			symbol++
			code_lengths[symbol] = length
			repeat--
		}
      }
    }
  }
  ok = 1

End:
  VP8LHuffmanTablesDeallocate(&tables)
  if !ok { return VP8LSetError(dec, VP8_STATUS_BITSTREAM_ERROR)  }
  return ok
}

// 'code_lengths' is pre-allocated temporary buffer, used for creating Huffman
// tree.
func ReadHuffmanCode(int alphabet_size, /*const*/ dec *VP8LDecoder, /*const*/ code_lengths *int, /*const*/ table *HuffmanTables) int {
  ok := 0
  size := 0
  var br *VP8LBitReader = &decoder.br
  simple_code := VP8LReadBits(br, 1)

  // C: stdlib.Memset(code_lengths, 0, alphabet_size * sizeof(*code_lengths))

  if simple_code {  // Read symbols, codes & code lengths directly.
    num_symbols := VP8LReadBits(br, 1) + 1
    first_symbol_len_code := VP8LReadBits(br, 1)
    // The first code is either 1 bit or 8 bit code.
    symbol := VP8LReadBits(br, tenary.If(first_symbol_len_code == 0, 1, 8))
    code_lengths[symbol] = 1
    // The second code (if present), is always 8 bits long.
    if num_symbols == 2 {
      symbol = VP8LReadBits(br, 8)
      code_lengths[symbol] = 1
    }
    ok = 1
  } else {  // Decode Huffman-coded code lengths.
    var i int
    var code_length_code_lengths [NUM_CODE_LENGTH_CODES]int = [NUM_CODE_LENGTH_CODES]int{0}
    num_codes := VP8LReadBits(br, 4) + 4
    assert.Assert(num_codes <= NUM_CODE_LENGTH_CODES)

    for i = 0; i < num_codes; i++ {
      code_length_code_lengths[kCodeLengthCodeOrder[i]] = VP8LReadBits(br, 3)
    }
    ok = ReadHuffmanCodeLengths(dec, code_length_code_lengths, alphabet_size, code_lengths)
  }

  ok = ok && !br.eos
  if ok {
    var bounded_code_lengths *int = // C: code_lengths // bidi index -> alphabet_size * sizeof(int)
    size = VP8LBuildHuffmanTable(table, HUFFMAN_TABLE_BITS, bounded_code_lengths, alphabet_size)
  }
  if !ok || size == 0 {
    return VP8LSetError(dec, VP8_STATUS_BITSTREAM_ERROR)
  }
  return size
}

func ReadHuffmanCodes(/* const */ dec *VP8LDecoder, xsize int, ysize int, color_cache_bits int, allow_recursion int) int {
  var i int
  var br *VP8LBitReader = &decoder.br
  var hdr *VP8LMetadata = &decoder.hdr
  huffman_image *uint32 = nil
  htree_groups *HTreeGroup = nil
  huffman_tables *HuffmanTables = &hdr.huffman_tables
  num_htree_groups := 1
  num_htree_groups_max := 1
  mapping *int = nil
  ok := 0

  // Check the table has been 0 initialized (through InitMetadata).
  assert.Assert(huffman_tables.root.start == nil)
  assert.Assert(huffman_tables.curr_segment == nil)

  if allow_recursion && VP8LReadBits(br, 1) {
    // use meta Huffman codes.
    huffman_precision := MIN_HUFFMAN_BITS + VP8LReadBits(br, NUM_HUFFMAN_BITS)
    huffman_xsize := VP8LSubSampleSize(xsize, huffman_precision)
    huffman_ysize := VP8LSubSampleSize(ysize, huffman_precision)
    huffman_pixs := huffman_xsize * huffman_ysize
    if !DecodeImageStream(huffman_xsize, huffman_ysize, /*is_level0=*/0, dec, &huffman_image) {
      goto Error
    }
    hdr.huffman_subsample_bits = huffman_precision
    for i = 0; i < huffman_pixs; i++ {
      // The huffman data is stored in red and green bytes.
      group := (huffman_image[i] >> 8) & 0xffff
      huffman_image[i] = group
      if group >= num_htree_groups_max {
        num_htree_groups_max = group + 1
      }
    }
    // Check the validity of num_htree_groups_max. If it seems too big, use a
    // smaller value for later. This will prevent big memory allocations to end
    // up with a bad bitstream anyway.
    // The value of 1000 is totally arbitrary. We know that num_htree_groups_max
    // is smaller than (1 << 16) and should be smaller than the number of pixels
    // (though the format allows it to be bigger).
    if num_htree_groups_max > 1000 || num_htree_groups_max > xsize * ysize {
      // Create a mapping from the used indices to the minimal set of used
      // values [0, num_htree_groups)
    //   mapping = (*int)WebPSafeMalloc(num_htree_groups_max, sizeof(*mapping));
    //   if (mapping == nil) {
    //     VP8LSetError(dec, VP8_STATUS_OUT_OF_MEMORY);
    //     goto Error;
    //   }
	  mapping := make([]int, num_htree_groups_max)
      // -1 means a value is unmapped, and therefore unused in the Huffman
      // image.
      // C: stdlib.Memset(mapping, 0xff, num_htree_groups_max * sizeof(*mapping))
      for num_htree_groups = 0, i = 0; i < huffman_pixs; i++ {
        // Get the current mapping for the group and remap the Huffman image.
        var mapped_group *int = &mapping[huffman_image[i]]
        if *mapped_group == -1 { *mapped_group = num_htree_groups++ }
        huffman_image[i] = *mapped_group
      }
    } else {
      num_htree_groups = num_htree_groups_max
    }
  }

  if br.eos { goto Error }

  if !ReadHuffmanCodesHelper(color_cache_bits, num_htree_groups, num_htree_groups_max, mapping, dec, huffman_tables, &htree_groups) {
    goto Error
  }
  ok = 1

  // All OK. Finalize pointers.
  hdr.huffman_image = huffman_image
  hdr.num_htree_groups = num_htree_groups
  hdr.htree_groups = htree_groups

Error:
  if !ok {
    VP8LHuffmanTablesDeallocate(huffman_tables)
    htree_groups = nil
  }
  return ok
}

// Helper function for reading the different Huffman codes and storing them in
// 'huffman_tables' and 'htree_groups'.
// If mapping is nil 'num_htree_groups_max' must equal 'num_htree_groups'.
// If it is not nil, it maps 'num_htree_groups_max' indices to the
// 'num_htree_groups' groups. If 'num_htree_groups_max' > 'num_htree_groups',
// some of those indices map to -1. This is used for non-balanced codes to
// limit memory usage.
func ReadHuffmanCodesHelper(int color_cache_bits, num_htree_groups int, num_htree_groups_max int, /*const*/ mapping *int, /*const*/ dec *VP8LDecoder, /*const*/ huffman_tables *HuffmanTables, *HTreeGroup* const htree_groups) int {
  int i, j, ok = 0, 0, 0
  max_alphabet_size := kAlphabetSize[0] + tenary.If((color_cache_bits > 0), 1 << color_cache_bits, 0)
  table_size := kTableSize[color_cache_bits]

  if ((mapping == nil && num_htree_groups != num_htree_groups_max) ||
      num_htree_groups > num_htree_groups_max) {
    goto Error
  }

//   code_lengths = (*int)WebPSafeCalloc((uint64)max_alphabet_size, sizeof(*code_lengths));
  code_lengths := make([]int, max_alphabet_size)
  
  *htree_groups = VP8LHtreeGroupsNew(num_htree_groups)

  if (*htree_groups == nil || code_lengths == nil ||
      !VP8LHuffmanTablesAllocate(num_htree_groups * table_size, huffman_tables)) {
    VP8LSetError(dec, VP8_STATUS_OUT_OF_MEMORY)
    goto Error
  }

  for i = 0; i < num_htree_groups_max; i++ {
    // If the index "i" is unused in the Huffman image, just make sure the
    // coefficients are valid but do not store them.
    if mapping != nil && mapping[i] == -1 {
      for j = 0; j < HUFFMAN_CODES_PER_META_CODE; j++ {
        alphabet_size := kAlphabetSize[j]
        if j == 0 && color_cache_bits > 0 {
          alphabet_size += (1 << color_cache_bits)
        }
        // Passing in nil so that nothing gets filled.
        if !ReadHuffmanCode(alphabet_size, dec, code_lengths, nil) {
          goto Error
        }
      }
    } else {
      var htree_group *HTreeGroup = &(*htree_groups)[tenary.If(mapping == nil, i, mapping[i])]
      var htrees *HuffmanCode  = htree_group.htrees
      var size int
      total_size := 0
      is_trivial_literal := 1
      max_bits := 0
      for j = 0; j < HUFFMAN_CODES_PER_META_CODE; j++ {
        alphabet_size := kAlphabetSize[j]
        if j == 0 && color_cache_bits > 0 {
          alphabet_size += (1 << color_cache_bits)
        }
        size = ReadHuffmanCode(alphabet_size, dec, code_lengths, huffman_tables)
        htrees[j] = huffman_tables.curr_segment.curr_table
        if size == 0 {
          goto Error
        }
        if is_trivial_literal && kLiteralMap[j] == 1 {
          is_trivial_literal = (htrees[j].bits == 0)
        }
        total_size += htrees[j].bits
        huffman_tables.curr_segment.curr_table += size
        if j <= ALPHA {
          local_max_bits := code_lengths[0]
          var k int
          for k = 1; k < alphabet_size; k++ {
            if code_lengths[k] > local_max_bits {
              local_max_bits = code_lengths[k]
            }
          }
          max_bits += local_max_bits
        }
      }
      htree_group.is_trivial_literal = is_trivial_literal
      htree_group.is_trivial_code = 0
      if is_trivial_literal {
        red := htrees[RED][0].value
        blue := htrees[BLUE][0].value
        alpha := htrees[ALPHA][0].value
        htree_group.literal_arb = (uint32(alpha) << 24) | (red << 16) | blue
        if total_size == 0 && htrees[GREEN][0].value < NUM_LITERAL_CODES {
          htree_group.is_trivial_code = 1
          htree_group.literal_arb |= htrees[GREEN][0].value << 8
        }
      }
      htree_group.use_packed_table = !htree_group.is_trivial_code && (max_bits < HUFFMAN_PACKED_BITS)
      if htree_group.use_packed_table { BuildPackedTable(htree_group) }
    }
  }
  ok = 1

Error:
  if !ok {
    VP8LHuffmanTablesDeallocate(huffman_tables)
    htree_groups = nil
  }
  return ok
}

//------------------------------------------------------------------------------
// Scaling.

// C: #if !defined(WEBP_REDUCE_SIZE)
func AllocateAndInitRescaler(/* const */ dec *VP8LDecoder, /*const*/ io *VP8Io) int {
  num_channels := 4
  in_width := io.mb_w
  out_width := io.scaled_width
  in_height := io.mb_h
  out_height := io.scaled_height
  work_size := 2 * num_channels * uint64(out_width)
  var work *rescaler_t  // Rescaler work area.
  scaled_data_size := uint64(out_width)
  var scaled_data *uint32 // Temporary storage for scaled BGRA data.
  // C: memory_size := sizeof(*dec.rescaler) +
                               // C: work_size * sizeof(*work) +
                               // C: scaled_data_size * sizeof(*scaled_data)
//   var memory *uint8 = (*uint8)WebPSafeMalloc(memory_size, sizeof(*memory));
//   if (memory == nil) {
//     return VP8LSetError(dec, VP8_STATUS_OUT_OF_MEMORY);
//   }
  memory := make([]uint8, memory_size)

  assert.Assert(dec.rescaler_memory == nil)
  dec.rescaler_memory = memory

  dec.rescaler = (*WebPRescaler)memory
  // C: memory += sizeof(*dec.rescaler)
  work = (rescaler_t*)memory
  // C: memory += work_size * sizeof(*work)
  scaled_data = (*uint32)memory

  if !WebPRescalerInit(dec.rescaler, in_width, in_height, (*uint8)scaled_data, out_width, out_height, 0, num_channels, work) {
    return 0
  }
  return 1
}
// C: #endif  // WEBP_REDUCE_SIZE

//------------------------------------------------------------------------------
// Export to ARGB

// C: #if !defined(WEBP_REDUCE_SIZE)

// We have special "export" function since we need to convert from BGRA
func Export(/* const */ rescaler *WebPRescaler, WEBP_CSP_MODE colorspace, rgba_stride int, /*const*/ rgba *uint8) int {
  var src *uint32 = (*uint32)rescaler.dst
  dst *uint8 = rgba
  dst_width := rescaler.dst_width
  num_lines_out := 0
  for (WebPRescalerHasPendingOutput(rescaler)) {
    WebPRescalerExportRow(rescaler)
    WebPMultARGBRow(src, dst_width, 1)
    VP8LConvertFromBGRA(src, dst_width, colorspace, dst)
    dst += rgba_stride
    num_lines_out++
  }
  return num_lines_out
}

// Emit scaled rows.
func EmitRescaledRowsRGBA(/* const */ dec *VP8LDecoder, in *uint8, in_stride int, mb_h int, /*const*/ out *uint8, out_stride int) int {
  var colorspace WEBP_CSP_MODE = dec.output.colorspace
  num_lines_in := 0
  num_lines_out := 0
  for num_lines_in < mb_h {
    var row_in *uint8 = in + ptrdiff_t(num_lines_in) * in_stride
    var row_out *uint8 = out + ptrdiff_t(num_lines_out) * out_stride
    lines_left := mb_h - num_lines_in
    needed_lines := WebPRescaleNeededLines(dec.rescaler, lines_left)
    var lines_imported int
    assert.Assert(needed_lines > 0 && needed_lines <= lines_left)
    WebPMultARGBRows(row_in, in_stride, dec.rescaler.src_width, needed_lines, 0)
    lines_imported = WebPRescalerImport(dec.rescaler, lines_left, row_in, in_stride)
    assert.Assert(lines_imported == needed_lines)
    num_lines_in += lines_imported
    num_lines_out += Export(dec.rescaler, colorspace, out_stride, row_out)
  }
  return num_lines_out
}

// C: #endif  // WEBP_REDUCE_SIZE

// Emit rows without any scaling.
func EmitRows(WEBP_CSP_MODE colorspace, /*const*/ row_in *uint8, in_stride int, mb_w int, mb_h int, /*const*/ out *uint8, out_stride int) int {
  lines := mb_h
  row_out *uint8 = out
  for lines-- > 0 {
    VP8LConvertFromBGRA((/* const */ *uint32)row_in, mb_w, colorspace, row_out)
    row_in += in_stride
    row_out += out_stride
  }
  return mb_h;  // Num rows out == num rows in.
}

//------------------------------------------------------------------------------
// Export to YUVA

func ConvertToYUVA(/* const */ src *uint32, width int, y_pos int, /*const*/ output *WebPDecBuffer) {
  var buf *WebPYUVABuffer = &output.u.YUVA

  // first, the luma plane
  WebPConvertARGBToY(src, buf.y + ptrdiff_t(y_pos) * buf.y_stride, width)

  // then U/V planes
  {
    var u *uint8 = buf.u + ptrdiff_t(y_pos >> 1) * buf.u_stride
    var v *uint8 = buf.v + ptrdiff_t(y_pos >> 1) * buf.v_stride
    // even lines: store values
    // odd lines: average with previous values
    WebPConvertARGBToUV(src, u, v, width, !(y_pos & 1))
  }
  // Lastly, store alpha if needed.
  if buf.a != nil {
    var a *uint8 = buf.a + ptrdiff_t(y_pos) * buf.a_stride
if constants.WORDS_BIGENDIAN {
    WebPExtractAlpha((*uint8)src + 0, 0, width, 1, a, 0)
} else {
    WebPExtractAlpha((*uint8)src + 3, 0, width, 1, a, 0)
}
  }
}

func ExportYUVA(/* const */ dec *VP8LDecoder, y_pos int) int {
  var rescaler *WebPRescaler = decoder.rescaler
  var src *uint32 = (*uint32)rescaler.dst
  dst_width := rescaler.dst_width
  num_lines_out := 0
  for WebPRescalerHasPendingOutput(rescaler) {
    WebPRescalerExportRow(rescaler)
    WebPMultARGBRow(src, dst_width, 1)
    ConvertToYUVA(src, dst_width, y_pos, dec.output)
    y_pos++
    num_lines_out++
  }
  return num_lines_out
}

func EmitRescaledRowsYUVA(/* const */ dec *VP8LDecoder, in *uint8, in_stride, mb_h int) int {
  num_lines_in := 0
  y_pos := dec.last_out_row
  for num_lines_in < mb_h {
    lines_left := mb_h - num_lines_in
    needed_lines := WebPRescaleNeededLines(dec.rescaler, lines_left)
    var lines_imported int
    WebPMultARGBRows(in, in_stride, dec.rescaler.src_width, needed_lines, 0)
    lines_imported = WebPRescalerImport(dec.rescaler, lines_left, in, in_stride)
    assert.Assert(lines_imported == needed_lines)
    num_lines_in += lines_imported
    in += ptrdiff_t(needed_lines) * in_stride
    y_pos += ExportYUVA(dec, y_pos)
  }
  return y_pos
}

// Returns true if alpha[] has non-0xff values.
func CheckNonOpaque(/* const */ alpha *uint8, width, height, y_step int) int {
  WebPInitAlphaProcessing()
  for ; height-- > 0; alpha += y_step {
    if WebPHasAlpha8b(alpha, width) { return 1  }
  }
  return 0
}

func EmitRowsYUVA(/* const  */in *uint8, /* const  */io *VP8Io, in_stride int, tmp_rgb *uint16, dec *VP8LDecoder) int {
  y_pos := dec.last_out_row
  width := io.mb_w
  num_rows := io.mb_h
  y_pos_final := y_pos + num_rows
  y_stride := dec.output.u.YUVA.y_stride
  uv_stride := dec.output.u.YUVA.u_stride
  a_stride := dec.output.u.YUVA.a_stride
  dst_a *uint8 = dec.output.u.YUVA.a
  dst_y *uint8 = dec.output.u.YUVA.y + ptrdiff_t(y_pos) * y_stride
  dst_u *uint8 = dec.output.u.YUVA.u + ptrdiff_t(y_pos >> 1) * uv_stride
  dst_v *uint8 = dec.output.u.YUVA.v + ptrdiff_t(y_pos >> 1) * uv_stride
  var r_ptr *uint8 = in + CHANNEL_OFFSET(1)
  var g_ptr *uint8 = in + CHANNEL_OFFSET(2)
  var b_ptr *uint8 = in + CHANNEL_OFFSET(3)
  var a_ptr *uint8 = nil
  has_alpha := 0

  // Make sure the lines are processed two by two from the start.
  assert.Assert(y_pos % 2 == 0)

  // Make sure num_rows is even. y_pos_final will check if it not.
  num_rows &= ~1

  if dst_a {
    dst_a += ptrdiff_t(y_pos) * a_stride
    a_ptr = in + CHANNEL_OFFSET(0)
    has_alpha = CheckNonOpaque(a_ptr, width, num_rows, in_stride)
  }
  // Process pairs of lines.
  WebPImportYUVAFromRGBA(r_ptr, g_ptr, b_ptr, a_ptr, /*step=*/4, in_stride, has_alpha, width, num_rows, tmp_rgb, y_stride, uv_stride, a_stride, dst_y, dst_u, dst_v, dst_a)

  y_pos += num_rows
  if y_pos_final == io.crop_bottom - io.crop_top && y_pos < y_pos_final {
    assert.Assert(y_pos + 1 == y_pos_final)
    // If we output the last line of an image with odd height.
    dst_y += ptrdiff_t(num_rows) * y_stride
    dst_u += ptrdiff_t(num_rows >> 1) * uv_stride
    dst_v += ptrdiff_t(num_rows >> 1) * uv_stride
    r_ptr += ptrdiff_t(num_rows) * in_stride
    g_ptr += ptrdiff_t(num_rows) * in_stride
    b_ptr += ptrdiff_t(num_rows) * in_stride
    if dst_a {
      dst_a += ptrdiff_t(num_rows) * a_stride
      a_ptr += ptrdiff_t(num_rows) * in_stride
      has_alpha = CheckNonOpaque(a_ptr, width, /*height=*/1, in_stride)
    }
    WebPImportYUVAFromRGBALastLine(r_ptr, g_ptr, b_ptr, a_ptr, /*step=*/4, has_alpha, width, tmp_rgb, dst_y, dst_u, dst_v, dst_a)
    y_pos = y_pos_final
  }
  return y_pos
}

//------------------------------------------------------------------------------
// Cropping.

// Sets io.mb_y, io.mb_h & io.mb_w according to start row, end row and
// crop options. Also updates the input data pointer, so that it points to the
// start of the cropped window. Note that pixels are in ARGB format even if
// 'in_data' is *uint8.
// Returns true if the crop window is not empty.
func SetCropWindow(/* const */ io *VP8Io,  y_start int, y_end int, in_data *uint8, pixel_stride int) int {
  assert.Assert(y_start < y_end)
  assert.Assert(io.crop_left < io.crop_right)
  if y_end > io.crop_bottom {
    y_end = io.crop_bottom;  // make sure we don't overflow on last row.
  }
  if y_start < io.crop_top {
    delta := io.crop_top - y_start
    y_start = io.crop_top
    *in_data += ptrdiff_t(delta) * pixel_stride
  }
  if y_start >= y_end {
    return 0  // Crop window is empty.
}

  // C: *in_data += io.crop_left * sizeof(uint32)

  io.mb_y = y_start - io.crop_top
  io.mb_w = io.crop_right - io.crop_left
  io.mb_h = y_end - y_start
  return 1;  // Non-empty crop window.
}

//------------------------------------------------------------------------------

func GetMetaIndex(/* const */ image *uint32, xsize int, bits int, x int, y int) int {
  if bits == 0 { return 0  }
  return image[xsize * (y >> bits) + (x >> bits)]
}

// C: static  GetHtreeGroupForPos *HTreeGroup(/* const */ hdr *VP8LMetadata, x int, y int) {
  meta_index := GetMetaIndex(hdr.huffman_image, hdr.huffman_xsize, hdr.huffman_subsample_bits, x, y)
  assert.Assert(meta_index < hdr.num_htree_groups)
  return hdr.htree_groups + meta_index
}

//------------------------------------------------------------------------------
// Main loop, with custom row-processing function

// If 'wait_for_biggest_batch' is true, wait for enough data to fill the
// argb_cache as much as possible (usually NUM_ARGB_CACHE_ROWS).
// C: typedef func (*ProcessRowsFunc)(/* const */ dec *VP8LDecoder, row int, wait_for_biggest_batch int)

func ApplyInverseTransforms(/* const */ dec *VP8LDecoder, start_row int, num_rows int , /*const*/ rows *uint32) {
  n := dec.next_transform
  cache_pixs := dec.width * num_rows
  end_row := start_row + num_rows
  var rows_in *uint32 = rows
  var rows_out *uint32 = decoder.argb_cache

  // Inverse transforms.
  for n-- > 0 {
    var transform *VP8LTransform = &decoder.transforms[n]
    VP8LInverseTransform(transform, start_row, end_row, rows_in, rows_out)
    rows_in = rows_out
  }
  if rows_in != rows_out {
    // No transform called, hence just copy.
    // C: stdlib.MemCpy(rows_out, rows_in, cache_pixs * sizeof(*rows_out))
  }
}

// Processes (transforms, scales & color-converts) the rows decoded after the
// last call.
func ProcessRows(/* const */ dec *VP8LDecoder, row int, wait_for_biggest_batch int) {
  var rows *uint32 = decoder.pixels + decoder.width * decoder.last_row
  num_rows int 

  // In case of YUV conversion and if we do not need to get to the last row.
  if wait_for_biggest_batch {
    // In case of YUV conversion, and if we do not use the whole cropping
    // region.
    if (!WebPIsRGBMode(dec.output.colorspace) && row >= dec.io.crop_top &&
        row < dec.io.crop_bottom) {
      // Make sure the number of rows to process is even.
      if row - dec.io.crop_top % 2 != 0 { return }
      // Make sure the cache is as full as possible.
      if (row % NUM_ARGB_CACHE_ROWS != 0 &&
          (row + 1) % NUM_ARGB_CACHE_ROWS != 0) {
        return
      }
    } else {
      if row % NUM_ARGB_CACHE_ROWS != 0 { return }
    }
  }
  num_rows = row - dec.last_row
  assert.Assert(row <= dec.io.crop_bottom)
  // We can't process more than NUM_ARGB_CACHE_ROWS at a time (that's the size
  // of argb_cache), but we currently don't need more than that.
  assert.Assert(num_rows <= NUM_ARGB_CACHE_ROWS)
  if num_rows > 0 {  // Emit output.
    var io *VP8Io = decoder.io
    rows_data *uint8 = (*uint8)dec.argb_cache
    // C: in_stride := io.width * sizeof(uint32);  // in unit of RGBA
    ApplyInverseTransforms(dec, dec.last_row, num_rows, rows)
    if !SetCropWindow(io, dec.last_row, row, &rows_data, in_stride) {
      // Nothing to output (this time).
    } else {
      var output *WebPDecBuffer = decoder.output
      if WebPIsRGBMode(output.colorspace) {  // convert to RGBA
        var buf *WebPRGBABuffer = &output.u.RGBA
        var rgba *uint8 = buf.rgba + ptrdiff_t(dec.last_out_row) * buf.stride
        num_rows_out := tenary.If(io.use_scaling, EmitRescaledRowsRGBA(dec, rows_data, in_stride, io.mb_h, rgba, buf.stride), EmitRows(output.colorspace, rows_data, in_stride, io.mb_w, io.mb_h, rgba, buf.stride))
        // Update 'last_out_row'.
        dec.last_out_row += num_rows_out
      } else {  // convert to YUVA
        dec.last_out_row = tenary.If(io.use_scaling, EmitRescaledRowsYUVA(dec, rows_data, in_stride, io.mb_h), EmitRowsYUVA(rows_data, io, in_stride, dec.accumulated_rgb_pixels, dec))
      }
      assert.Assert(dec.last_out_row <= output.height)
    }
  }

  // Update 'last_row'.
  dec.last_row = row
  assert.Assert(dec.last_row <= dec.height)
}

// Row-processing for the special case when alpha data contains only one
// transform (color indexing), and trivial non-green literals.
func Is8bOptimizable(/* const */ hdr *VP8LMetadata) int {
  var i int
  if hdr.color_cache_size > 0 { return 0  }
  // When the Huffman tree contains only one symbol, we can skip the
  // call to ReadSymbol() for red/blue/alpha channels.
  for i = 0; i < hdr.num_htree_groups; i++ {
    var htrees *HuffmanCode = hdr.htree_groups[i].htrees
    if htrees[RED][0].bits > 0 { return 0  }
    if htrees[BLUE][0].bits > 0 { return 0  }
    if htrees[ALPHA][0].bits > 0 { return 0  }
  }
  return 1
}

func AlphaApplyFilter(/* const */ alph_dec *ALPHDecoder, first_row int, last_row int, out *uint8, stride int) {
  if alph_dec.filter != WEBP_FILTER_NONE {
    var y int
    var prev_line *uint8 = alph_dec.prev_line
    assert.Assert(WebPUnfilters[alph_dec.filter] != nil)
    for y = first_row; y < last_row; y++ {
      WebPUnfilters[alph_dec.filter](prev_line, out, out, stride)
      prev_line = out
      out += stride
    }
    alph_dec.prev_line = prev_line
  }
}

func ExtractPalettedAlphaRows(/* const */ dec *VP8LDecoder, last_row int) {
  // For vertical and gradient filtering, we need to decode the part above the
  // crop_top row, in order to have the correct spatial predictors.
  var alph_dec *ALPHDecoder = (*ALPHDecoder)dec.io.opaque
  top_row := (alph_dec.filter == WEBP_FILTER_NONE ||
                       alph_dec.filter == WEBP_FILTER_HORIZONTAL)
                          ? dec.io.crop_top
                          : dec.last_row
  first_row := (dec.last_row < top_row) ? top_row : dec.last_row
  assert.Assert(last_row <= dec.io.crop_bottom)
  if last_row > first_row {
    // Special method for paletted alpha data. We only process the cropped area.
    width := dec.io.width
    out *uint8 = alph_dec.output + width * first_row
    var in *uint8 = (*uint8)dec.pixels + dec.width * first_row
    var transform *VP8LTransform = &decoder.transforms[0]
    assert.Assert(dec.next_transform == 1)
    assert.Assert(transform.type == COLOR_INDEXING_TRANSFORM)
    VP8LColorIndexInverseTransformAlpha(transform, first_row, last_row, in, out)
    AlphaApplyFilter(alph_dec, first_row, last_row, out, width)
  }
  dec.last_row = dec.last_out_row = last_row
}

//------------------------------------------------------------------------------
// Helper functions for fast pattern copy (8b and 32b)

// cyclic rotation of pattern word
func Rotate8b(uint32 V) uint32 {
	if constants.WORDS_BIGENDIAN {
	return ((V & uint(0xff000000)) >> 24) | (V << 8)
	} else {
	return ((V & uint(0xff)) << 24) | (V >> 8)
	}
}

// copy 1, 2 or 4-bytes pattern
func CopySmallPattern8b(/* const */ src *uint8, dst *uint8, length int, uint32 pattern) {
  var i int
  // align 'dst' to 4-bytes boundary. Adjust the pattern along the way.
  for (uintptr_t)dst & 3 {
    *dst++ = *src++
    pattern = Rotate8b(pattern)
    length--
  }
  // Copy the pattern 4 bytes at a time.
  for i = 0; i < (length >> 2); i++ {
    ((*uint32)dst)[i] = pattern
  }
  // Finish with left-overs. 'pattern' is still correctly positioned, // so no Rotate8b() call is needed.
  for i <<= 2; i < length; i++ {
    dst[i] = src[i]
  }
}

func CopyBlock8b(/* const */ dst *uint8, dist int, length int) {
  var src *uint8 = dst - dist
  if length >= 8 {
    pattern := 0
    switch dist {
      case 1:
        pattern = src[0]
// C: #if defined(__arm__) || defined(_M_ARM)  // arm doesn't like multiply that much
        pattern |= pattern << 8
        pattern |= pattern << 16
// C: #elif defined(WEBP_USE_MIPS_DSP_R2)
        __asm__ volatile("replv.qb %0, %0" : "+r"(pattern))
// C: #else
        pattern = uint(0x01010101) * pattern
// C: #endif
        break
      case 2:
if !constants.WORDS_BIGENDIAN {
        // C: stdlib.MemCpy(&pattern, src, sizeof(uint16))
} else {
        pattern = (uint32(src[0]) << 8) | src[1]
	  }
// C: #if defined(__arm__) || defined(_M_ARM)
        pattern |= pattern << 16
// C: #elif defined(WEBP_USE_MIPS_DSP_R2)
        __asm__ volatile("replv.ph %0, %0" : "+r"(pattern))
// C: #else
        pattern = uint(0x00010001) * pattern
// C: #endif
        break
      case 4:
        // C: stdlib.MemCpy(&pattern, src, sizeof(uint32))
        break
      default:
        goto Copy
    }
    CopySmallPattern8b(src, dst, length, pattern)
    return
  }
Copy:
  if dist >= length {  // no overlap . use stdlib.MemCpy()
    // C: stdlib.MemCpy(dst, src, length * sizeof(*dst))
  } else {
    var i int
    for i = 0; i < length; i++ dst[i] = src[i];

  }
}

// copy pattern of 1 or 2 uint32's
func CopySmallPattern32b(/* const */ src *uint32, dst *uint32, length int, uint64 pattern) {
  var i int
  if (uintptr_t)dst & 4 {  // Align 'dst' to 8-bytes boundary.
    *dst++ = *src++
    pattern = (pattern >> 32) | (pattern << 32)
    length--
  }
  assert.Assert(0 == ((uintptr_t)dst & 7))
  for i = 0; i < (length >> 1); i++ {
    ((*uint64)dst)[i] = pattern;  // Copy the pattern 8 bytes at a time.
  }
  if length & 1 {  // Finish with left-over.
    dst[i << 1] = src[i << 1]
  }
}

func CopyBlock32b(/* const */ dst *uint32, dist int, length int) {
  var src *uint32 = dst - dist
  if dist <= 2 && length >= 4 && ((uintptr_t)dst & 3) == 0 {
    var pattern uint64
    if dist == 1 {
      pattern = uint64(src[0])
      pattern |= pattern << 32
    } else {
      // C: stdlib.MemCpy(&pattern, src, sizeof(pattern))
    }
    CopySmallPattern32b(src, dst, length, pattern)
  } else if dist >= length {  // no overlap
    // C: stdlib.MemCpy(dst, src, length * sizeof(*dst))
  } else {
    var i int
    for i = 0; i < length; i++ dst[i] = src[i];

  }
}

//------------------------------------------------------------------------------

func DecodeAlphaData(/* const */ dec *VP8LDecoder, /*const*/ data *uint8, width, height int, last_row int) int {
  ok := 1
  row := dec.last_pixel / width
  col := dec.last_pixel % width
  var br *VP8LBitReader = &decoder.br
  var hdr *VP8LMetadata = &decoder.hdr
  pos := dec.last_pixel;          // current position
  end := width * height;     // End of data
  last := width * last_row;  // Last pixel to decode
  len_code_limit := NUM_LITERAL_CODES + NUM_LENGTH_CODES
  mask := hdr.huffman_mask
  const htree_group *HTreeGroup = (pos < last) ? GetHtreeGroupForPos(hdr, col, row) : nil
  assert.Assert(pos <= end)
  assert.Assert(last_row <= height)
  assert.Assert(Is8bOptimizable(hdr))

  for !br.eos && pos < last {
    var code int
    // Only update when changing tile.
    if (col & mask) == 0 {
      htree_group = GetHtreeGroupForPos(hdr, col, row)
    }
    assert.Assert(htree_group != nil)
    VP8LFillBitWindow(br)
    code = ReadSymbol(htree_group.htrees[GREEN], br)
    if code < NUM_LITERAL_CODES {  // Literal
      data[pos] = code
      pos++
      col++
      if col >= width {
        col = 0
        row++
        if row <= last_row && (row % NUM_ARGB_CACHE_ROWS == 0) {
          ExtractPalettedAlphaRows(dec, row)
        }
      }
    } else if code < len_code_limit {  // Backward reference
      var dist_code, dist int
      length_sym := code - NUM_LITERAL_CODES
      length := GetCopyLength(length_sym, br)
      dist_symbol := ReadSymbol(htree_group.htrees[DIST], br)
      VP8LFillBitWindow(br)
      dist_code = GetCopyDistance(dist_symbol, br)
      dist = PlaneCodeToDistance(width, dist_code)
      if pos >= dist && end - pos >= length {
        CopyBlock8b(data + pos, dist, length)
      } else {
        ok = 0
        goto End
      }
      pos += length
      col += length
      for col >= width {
        col -= width
        row++
        if row <= last_row && (row % NUM_ARGB_CACHE_ROWS == 0) {
          ExtractPalettedAlphaRows(dec, row)
        }
      }
      if pos < last && (col & mask) {
        htree_group = GetHtreeGroupForPos(hdr, col, row)
      }
    } else {  // Not reached
      ok = 0
      goto End
    }
    br.eos = VP8LIsEndOfStream(br)
  }
  // Process the remaining rows corresponding to last row-block.
  ExtractPalettedAlphaRows(dec, row > last_row ? last_row : row)

End:
  br.eos = VP8LIsEndOfStream(br)
  if !ok || (br.eos && pos < end) {
    return VP8LSetError(
        dec, tenary.If(br.eos, VP8_STATUS_SUSPENDED, VP8_STATUS_BITSTREAM_ERROR))
  }
  dec.last_pixel = pos
  return ok
}

func SaveState(/* const */ dec *VP8LDecoder, last_pixel int) {
  assert.Assert(dec.incremental)
  dec.saved_br = dec.br
  dec.saved_last_pixel = last_pixel
  if dec.hdr.color_cache_size > 0 {
    VP8LColorCacheCopy(&dec.hdr.color_cache, &dec.hdr.saved_color_cache)
  }
}

func RestoreState(/* const */ dec *VP8LDecoder) {
  assert.Assert(dec.br.eos)
  dec.status = VP8_STATUS_SUSPENDED
  dec.br = dec.saved_br
  dec.last_pixel = dec.saved_last_pixel
  if dec.hdr.color_cache_size > 0 {
    VP8LColorCacheCopy(&dec.hdr.saved_color_cache, &dec.hdr.color_cache)
  }
}


func DecodeImageData(/* const */ dec *VP8LDecoder, /*const*/ data *uint32, width, height int, last_row int , process_func ProcessRowsFunc ) int {
  row := dec.last_pixel / width
  col := dec.last_pixel % width
  var br *VP8LBitReader = &decoder.br
  var hdr *VP8LMetadata = &decoder.hdr
  src *uint32 = data + dec.last_pixel
  last_cached *uint32 = src
  var src_end *uint32 = data + width * height;     // End of data
  var src_last *uint32 = data + width * last_row;  // Last pixel to decode
  len_code_limit := NUM_LITERAL_CODES + NUM_LENGTH_CODES
  color_cache_limit := len_code_limit + hdr.color_cache_size
  next_sync_row := dec.incremental ? row : 1 << 24
  var color_cache *VP8LColorCache = (hdr.color_cache_size > 0) ? &hdr.color_cache : nil
  mask := hdr.huffman_mask
  var htree_group *HTreeGroup = (src < src_last) ? GetHtreeGroupForPos(hdr, col, row) : nil
  assert.Assert(dec.last_row < last_row)
  assert.Assert(src_last <= src_end)

  for src < src_last {
    var code int
    if row >= next_sync_row {
      SaveState(dec, int(src - data))
      next_sync_row = row + SYNC_EVERY_N_ROWS
    }
    // Only update when changing tile. Note we could use this test:
    // if "((((prev_col ^ col) | prev_row ^ row)) > mask)" . tile changed
    // but that's actually slower and needs storing the previous col/row.
    if (col & mask) == 0 {
      htree_group = GetHtreeGroupForPos(hdr, col, row)
    }
    assert.Assert(htree_group != nil)
    if htree_group.is_trivial_code {
      *src = htree_group.literal_arb
      goto AdvanceByOne
    }
    VP8LFillBitWindow(br)
    if htree_group.use_packed_table {
      code = ReadPackedSymbols(htree_group, br, src)
      if VP8LIsEndOfStream(br) { break }
      if code == PACKED_NON_LITERAL_CODE { goto AdvanceByOne }
    } else {
      code = ReadSymbol(htree_group.htrees[GREEN], br)
    }
    if VP8LIsEndOfStream(br) { break }
    if code < NUM_LITERAL_CODES {  // Literal
      if htree_group.is_trivial_literal {
        *src = htree_group.literal_arb | (code << 8)
      } else {
        var red, blue, alpha int
        red = ReadSymbol(htree_group.htrees[RED], br)
        VP8LFillBitWindow(br)
        blue = ReadSymbol(htree_group.htrees[BLUE], br)
        alpha = ReadSymbol(htree_group.htrees[ALPHA], br)
        if VP8LIsEndOfStream(br) { break }
        *src = (uint32(alpha) << 24) | (red << 16) | (code << 8) | blue
      }
    AdvanceByOne:
      src++
      col++
      if col >= width {
        col = 0
        row++
        if process_func != nil {
          if row <= last_row {
            process_func(dec, row, /*wait_for_biggest_batch=*/1)
          }
        }
        if color_cache != nil {
          for last_cached < src {
            VP8LColorCacheInsert(color_cache, *last_cached++)
          }
        }
      }
    } else if code < len_code_limit {  // Backward reference
      var dist_code, dist int
      length_sym := code - NUM_LITERAL_CODES
      length := GetCopyLength(length_sym, br)
      dist_symbol := ReadSymbol(htree_group.htrees[DIST], br)
      VP8LFillBitWindow(br)
      dist_code = GetCopyDistance(dist_symbol, br)
      dist = PlaneCodeToDistance(width, dist_code)

      if VP8LIsEndOfStream(br) { break }
      if src - data < ptrdiff_t(dist) || src_end - src < ptrdiff_t(length) {
        goto Error
      } else {
        CopyBlock32b(src, dist, length)
      }
      src += length
      col += length
      for col >= width {
        col -= width
        row++
        if process_func != nil {
          if row <= last_row {
            process_func(dec, row, /*wait_for_biggest_batch=*/1)
          }
        }
      }
      // Because of the check done above (before 'src' was incremented by
      // 'length'), the following holds true.
      assert.Assert(src <= src_end)
      if col & mask { htree_group = GetHtreeGroupForPos(hdr, col, row) }
      if color_cache != nil {
        for last_cached < src {
          VP8LColorCacheInsert(color_cache, *last_cached++)
        }
      }
    } else if code < color_cache_limit {  // Color cache
      key := code - len_code_limit
      assert.Assert(color_cache != nil)
      for last_cached < src {
        VP8LColorCacheInsert(color_cache, *last_cached++)
      }
      *src = VP8LColorCacheLookup(color_cache, key)
      goto AdvanceByOne
    } else {  // Not reached
      goto Error
    }
  }

  br.eos = VP8LIsEndOfStream(br)
  // In incremental decoding:
  // br.eos && src < src_last: if 'br' reached the end of the buffer and
  // 'src_last' has not been reached yet, there is not enough data. 'dec' has to
  // be reset until there is more data.
  // !br.eos && src < src_last: this cannot happen as either the buffer is
  // fully read, either enough has been read to reach 'src_last'.
  // src >= src_last: 'src_last' is reached, all is fine. 'src' can actually go
  // beyond 'src_last' in case the image is cropped and an LZ77 goes further.
  // The buffer might have been enough or there is some left. 'br.eos' does
  // not matter.
  assert.Assert(!dec.incremental || (br.eos && src < src_last) || src >= src_last)
  if dec.incremental && br.eos && src < src_last {
    RestoreState(dec)
  } else if (dec.incremental && src >= src_last) || !br.eos {
    // Process the remaining rows corresponding to last row-block.
    if process_func != nil {
      process_func(dec, row > last_row ? last_row : row, /*wait_for_biggest_batch=*/0)
    }
    dec.status = VP8_STATUS_OK
    dec.last_pixel = int(src - data);  // end-of-scan marker
  } else {
    // if not incremental, and we are past the end of buffer (eos=1), then this
    // is a real bitstream error.
    goto Error
  }
  return 1

Error:
  return VP8LSetError(dec, VP8_STATUS_BITSTREAM_ERROR)
}

// -----------------------------------------------------------------------------
// VP8LTransform

// For security reason, we need to remap the color map to span
// the total possible bundled values, and not just the num_colors.
func ExpandColorMap(num_colors int, /*const*/ transform *VP8LTransform) int {
  var i int
  final_num_colors := 1 << (8 >> transform.bits)
//   var new_color_map *uint32 = (*uint32)WebPSafeMalloc((uint64)final_num_colors, sizeof(*new_color_map));
//   if (new_color_map == nil) {
//     return 0;
//   } else {
	new_color_map := make([]uint32, final_num_colors)

    var data *uint8 = (*uint8)transform.data
    var new_data *uint8 = (*uint8)new_color_map
    new_color_map[0] = transform.data[0]
    for i = 4; i < 4 * num_colors; i++ {
      // Equivalent to VP8LAddPixels(), on a byte-basis.
      new_data[i] = (data[i] + new_data[i - 4]) & 0xff
    }
    for ; i < 4 * final_num_colors; i++ {
      new_data[i] = 0;  // black tail.
    }
	
    transform.data = new_color_map
//   }
  return 1
}

func ReadTransform(/* const */ xsize *int, ysize int *const, /*const*/ decoder *VP8LDecoder) int {
  ok := 1
  var br *VP8LBitReader = &decoder.br
  transform *VP8LTransform = &dec.transforms[dec.next_transform]
  var type VP8LImageTransformType = VP8LImageTransformType(VP8LReadBits)(br, 2)

  // Each transform type can only be present once in the stream.
  if dec.transforms_seen & (uint(1) << type) {
    return 0;  // Already there, let's not accept the second same transform.
  }
  dec.transforms_seen |= (uint(1) << type)

  transform.type = type
  transform.xsize = *xsize
  transform.ysize = *ysize
  transform.data = nil
  dec.next_transform++
  assert.Assert(dec.next_transform <= NUM_TRANSFORMS)

  switch type {
    case PREDICTOR_TRANSFORM:
    case CROSS_COLOR_TRANSFORM:
      transform.bits = MIN_TRANSFORM_BITS + VP8LReadBits(br, NUM_TRANSFORM_BITS)
      ok = DecodeImageStream(
          VP8LSubSampleSize(transform.xsize, transform.bits), VP8LSubSampleSize(transform.ysize, transform.bits), /*is_level0=*/0, dec, &transform.data)
      break
    case COLOR_INDEXING_TRANSFORM: {
      num_colors := VP8LReadBits(br, 8) + 1
      bits := (num_colors > 16)  ? 0
                       : (num_colors > 4) ? 1
                       : (num_colors > 2) ? 2
                                          : 3
      *xsize = VP8LSubSampleSize(transform.xsize, bits)
      transform.bits = bits
      ok = DecodeImageStream(num_colors, /*ysize=*/1, /*is_level0=*/0, dec, &transform.data)
      if ok && !ExpandColorMap(num_colors, transform) {
        return VP8LSetError(dec, VP8_STATUS_OUT_OF_MEMORY)
      }
      break
    }
    case SUBTRACT_GREEN_TRANSFORM:
      break
    default:
      assert.Assert(0);  // can't happen
      break
  }

  return ok
}

// -----------------------------------------------------------------------------
// VP8LMetadata

func InitMetadata(/* const */ hdr *VP8LMetadata) {
  assert.Assert(hdr != nil)
  // C: stdlib.Memset(hdr, 0, sizeof(*hdr))
}

func ClearMetadata(/* const */ hdr *VP8LMetadata) {
  assert.Assert(hdr != nil)

  VP8LHuffmanTablesDeallocate(&hdr.huffman_tables)
  hdr.htree_groups = nil
  VP8LColorCacheClear(&hdr.color_cache)
  VP8LColorCacheClear(&hdr.saved_color_cache)
  InitMetadata(hdr)
}

// -----------------------------------------------------------------------------
// VP8LDecoder

// Allocates and initialize a new lossless decoder instance.
func VP8LNew() *VP8LDecoder{
//   var dec *VP8LDecoder = (*VP8LDecoder)WebPSafeCalloc(uint64(1), sizeof(*dec));
//   if dec == nil { return nil }
	dec := &VP8LDecoder{
		status = VP8_STATUS_OK
		state = READ_DIM
	}

  VP8LDspInit();  // Init critical function pointers.

  return dec
}

// Resets the decoder in its initial state, reclaiming memory.
// Preserves the dec.status value.
func VP8LClear(/* const */ dec *VP8LDecoder) {
  if dec == nil { {return }}
  ClearMetadata(&dec.hdr)

  dec.pixels = nil
  dec.rescaler_memory = nil
  dec.output = nil;  // leave no trace behind
  
  dec.next_transform = 0
  dec.transforms_seen = 0
}

func UpdateDecoder(/* const */ dec *VP8LDecoder, width, height int) {
  var hdr *VP8LMetadata = &decoder.hdr
  num_bits := hdr.huffman_subsample_bits
  dec.width = width
  dec.height = height

  hdr.huffman_xsize = VP8LSubSampleSize(width, num_bits)
  hdr.huffman_mask = tenary.If(num_bits == 0, ~0, (1 << num_bits) - 1)
}

func DecodeImageStream(xsize int, ysize int, int is_level0, /*const*/ dec *VP8LDecoder, *uint32* const decoded_data) int {
  ok := 1
  transform_xsize := xsize
  transform_ysize := ysize
  var br *VP8LBitReader = &decoder.br
  var hdr *VP8LMetadata = &decoder.hdr
  data *uint32 = nil
  color_cache_bits := 0

  // Read the transforms (may recurse).
  if is_level0 {
    for ok && VP8LReadBits(br, 1) {
      ok = ReadTransform(&transform_xsize, &transform_ysize, dec)
    }
  }

  // Color cache
  if ok && VP8LReadBits(br, 1) {
    color_cache_bits = VP8LReadBits(br, 4)
    ok = (color_cache_bits >= 1 && color_cache_bits <= MAX_CACHE_BITS)
    if !ok {
      VP8LSetError(dec, VP8_STATUS_BITSTREAM_ERROR)
      goto End
    }
  }

  // Read the Huffman codes (may recurse).
  ok = ok && ReadHuffmanCodes(dec, transform_xsize, transform_ysize, color_cache_bits, is_level0)
  if !ok {
    VP8LSetError(dec, VP8_STATUS_BITSTREAM_ERROR)
    goto End
  }

  // Finish setting up the color-cache
  if color_cache_bits > 0 {
    hdr.color_cache_size = 1 << color_cache_bits
    if !VP8LColorCacheInit(&hdr.color_cache, color_cache_bits) {
      ok = VP8LSetError(dec, VP8_STATUS_OUT_OF_MEMORY)
      goto End
    }
  } else {
    hdr.color_cache_size = 0
  }
  UpdateDecoder(dec, transform_xsize, transform_ysize)

  if is_level0 {  // level 0 complete
    dec.state = READ_HDR
    goto End
  }

  {
    total_size := uint64(transform_xsize * transform_ysize)
    // data = (*uint32)WebPSafeMalloc(total_size, sizeof(*data));
    // if (data == nil) {
    //   ok = VP8LSetError(dec, VP8_STATUS_OUT_OF_MEMORY);
    //   goto End;
    // }
	data := make([]uint32, total_size)
  }

  // Use the Huffman trees to decode the LZ77 encoded data.
  ok = DecodeImageData(dec, data, transform_xsize, transform_ysize, transform_ysize, nil)
  ok = ok && !br.eos

End:
  if !ok {
	
    ClearMetadata(hdr)
  } else {
    if decoded_data != nil {
      *decoded_data = data
    } else {
      // We allocate image data in this function only for transforms. At level 0
      // (that is: not the transforms), we shouldn't have allocated anything.
      assert.Assert(data == nil)
      assert.Assert(is_level0)
    }
    dec.last_pixel = 0;  // Reset for future DECODE_DATA_FUNC() calls.
    if !is_level0 { ClearMetadata(hdr) }  // Clean up temporary data behind.
  }
  return ok
}

//------------------------------------------------------------------------------
// Allocate internal buffers dec.pixels and dec.argb_cache.
func AllocateInternalBuffers32b(/* const */ dec *VP8LDecoder, final_width int) int {
  num_pixels := uint64(dec.width) * dec.height
  // Scratch buffer corresponding to top-prediction row for transforming the
  // first row in the row-blocks. Not needed for paletted alpha.
  cache_top_pixels := uint16(final_width)
  // Scratch buffer for temporary BGRA storage. Not needed for paletted alpha.
  cache_pixels := uint64(final_width) * NUM_ARGB_CACHE_ROWS
  // Scratch buffer to accumulate RGBA values (hence 4*)for YUV conversion.
  accumulated_rgb_pixels := 0
  var total_num_pixels uint64
  if dec.output != nil && !WebPIsRGBMode(dec.output.colorspace) {
    uv_width := (dec.io.crop_right - dec.io.crop_left + 1) >> 1
    accumulated_rgb_pixels = // C: 4 * uv_width * sizeof(*dec.accumulated_rgb_pixels) / sizeof(uint32)
  }
  total_num_pixels = num_pixels + cache_top_pixels + cache_pixels + accumulated_rgb_pixels
  assert.Assert(dec.width <= final_width)
//   dec.pixels = (*uint32)WebPSafeMalloc(total_num_pixels, sizeof(uint32));
//   if (dec.pixels == nil) {
//     dec.argb_cache = nil;  // for soundness
//     return VP8LSetError(dec, VP8_STATUS_OUT_OF_MEMORY);
//   }
  dec.pixels = make([]uint32, total_num_pixels) // NOTE: have the feeling that this should be divided by 4

  dec.argb_cache = dec.pixels + num_pixels + cache_top_pixels
  dec.accumulated_rgb_pixels = accumulated_rgb_pixels == 0
          ? nil
          : (*uint16)(dec.pixels + num_pixels + cache_top_pixels +
                        cache_pixels)

  return 1
}

func AllocateInternalBuffers8b(/* const */ dec *VP8LDecoder) int {
  total_num_pixels := uint64(dec.width * dec.height;)
  dec.argb_cache = nil;  // for soundness
//   dec.pixels = (*uint32)WebPSafeMalloc(total_num_pixels, sizeof(uint8));
//   if (dec.pixels == nil) {
//     return VP8LSetError(dec, VP8_STATUS_OUT_OF_MEMORY);
//   }
  dec.pixels = make([]uint32, total_num_pixels) // NOTE: have the feeling that this should be divided by 4

  return 1
}

//------------------------------------------------------------------------------

// Special row-processing that only stores the alpha data.
func ExtractAlphaRows(/* const */ dec *VP8LDecoder, last_row int, wait_for_biggest_batch int) {
  cur_row := dec.last_row
  num_rows := last_row - cur_row
  var in *uint32 = decoder.pixels + decoder.width * cur_row

  if wait_for_biggest_batch && last_row % NUM_ARGB_CACHE_ROWS != 0 {
    return
  }
  assert.Assert(last_row <= dec.io.crop_bottom)
  for num_rows > 0 {
    num_rows_to_process := (num_rows > NUM_ARGB_CACHE_ROWS) ? NUM_ARGB_CACHE_ROWS : num_rows
    // Extract alpha (which is stored in the green plane).
    var alph_dec *ALPHDecoder = (*ALPHDecoder)dec.io.opaque
    var output *uint8 = alph_dec.output
    width := dec.io.width;  // the final width (!= dec.width)
    cache_pixs := width * num_rows_to_process
    var dst *uint8 = output + width * cur_row
    var src *uint32 = decoder.argb_cache
    ApplyInverseTransforms(dec, cur_row, num_rows_to_process, in)
    WebPExtractGreen(src, dst, cache_pixs)
    AlphaApplyFilter(alph_dec, cur_row, cur_row + num_rows_to_process, dst, width)
    num_rows -= num_rows_to_process
    in += num_rows_to_process * dec.width
    cur_row += num_rows_to_process
  }
  assert.Assert(cur_row == last_row)
  dec.last_row = dec.last_out_row = last_row
}

// Decodes image header for alpha data stored using lossless compression.
// Returns false in case of error.
func VP8LDecodeAlphaHeader(/* const */ alph_dec *ALPHDecoder, /*const*/ data *uint8, data_size uint64) int {
  ok := 0
  dec *VP8LDecoder = VP8LNew()

  if dec == nil { return 0  }

  assert.Assert(alph_dec != nil)

  dec.width = alph_dec.width
  dec.height = alph_dec.height
  dec.io = &alph_dec.io
  dec.io.opaque = alph_dec
  dec.io.width = alph_dec.width
  dec.io.height = alph_dec.height

  dec.status = VP8_STATUS_OK
  VP8LInitBitReader(&dec.br, data, data_size)

  if !DecodeImageStream(alph_dec.width, alph_dec.height, /*is_level0=*/1, dec, /*decoded_data=*/nil) {
    goto Err
  }

  // Special case: if alpha data uses only the color indexing transform and
  // doesn't use color cache (a frequent case), we will use DecodeAlphaData()
  // method that only needs allocation of 1 byte per pixel (alpha channel).
  if (dec.next_transform == 1 &&
      dec.transforms[0].type == COLOR_INDEXING_TRANSFORM &&
      Is8bOptimizable(&dec.hdr)) {
    alph_dec.use_8b_decode = 1
    ok = AllocateInternalBuffers8b(dec)
  } else {
    // Allocate internal buffers (note that dec.width may have changed here).
    alph_dec.use_8b_decode = 0
    ok = AllocateInternalBuffers32b(dec, alph_dec.width)
  }

  if !ok { goto Err }

  // Only set here, once we are sure it is valid (to afunc thread races).
  alph_dec.vp8l_dec = dec
  return 1

Err:
  return 0
}

// Decodes *at *least 'last_row' rows of alpha. If some of the initial rows are
// already decoded in previous call(s), it will resume decoding from where it
// was paused.
// Returns false in case of bitstream error.
func VP8LDecodeAlphaImageStream(/* const */ alph_dec *ALPHDecoder, last_row int) int {
  var dec *VP8LDecoder = alph_dec.vp8l_dec
  assert.Assert(dec != nil)
  assert.Assert(last_row <= dec.height)

  if dec.last_row >= last_row {
    return 1;  // done
  }

  if !alph_dec.use_8b_decode { WebPInitAlphaProcessing() }

  // Decode (with special row processing).
  return alph_dec.use_8b_decode
             ? DecodeAlphaData(dec, (*uint8)dec.pixels, dec.width, dec.height, last_row)
             : DecodeImageData(dec, dec.pixels, dec.width, dec.height, last_row, ExtractAlphaRows)
}

//------------------------------------------------------------------------------

// Decodes the image header. Returns false in case of error.
func VP8LDecodeHeader(/* const */ dec *VP8LDecoder, /* const */ io *VP8Io) int {
  var width, height, has_alpha int

  if dec == nil { return 0  }
  if io == nil {
    return VP8LSetError(dec, VP8_STATUS_INVALID_PARAM)
  }

  dec.io = io
  dec.status = VP8_STATUS_OK
  {
    var bounded_data *uint8 = io.data
    VP8LInitBitReader(&dec.br, bounded_data, io.data_size)
  }
  if !ReadImageInfo(&dec.br, &width, &height, &has_alpha) {
    VP8LSetError(dec, VP8_STATUS_BITSTREAM_ERROR)
    goto Error
  }
  dec.state = READ_DIM
  io.width = width
  io.height = height

  if !DecodeImageStream(width, height, /*is_level0=*/1, dec, /*decoded_data=*/nil) {
    goto Error
  }
  return 1

Error:
  VP8LClear(dec)
  assert.Assert(dec.status != VP8_STATUS_OK)
  return 0
}

// Decodes an image. It's required to decode the lossless header before calling
// this function. Returns false in case of error, with updated dec.status.
func VP8LDecodeImage(/* const */ dec *VP8LDecoder) int {
  io *VP8Io = nil
  params *WebPDecParams = nil

  if dec == nil { return 0  }

  assert.Assert(dec.hdr.huffman_tables.root.start != nil)
  assert.Assert(dec.hdr.htree_groups != nil)
  assert.Assert(dec.hdr.num_htree_groups > 0)

  io = dec.io
  assert.Assert(io != nil)
  params = (*WebPDecParams)io.opaque
  assert.Assert(params != nil)

  // Initialization.
  if dec.state != READ_DATA {
    dec.output = params.output
    assert.Assert(dec.output != nil)

    if !WebPIoInitFromOptions(params.options, io, MODE_BGRA) {
      VP8LSetError(dec, VP8_STATUS_INVALID_PARAM)
      goto Err
    }

    if !AllocateInternalBuffers32b(dec, io.width) { goto Err }

// C: #if !defined(WEBP_REDUCE_SIZE)
    if io.use_scaling && !AllocateAndInitRescaler(dec, io) { goto Err }
// C: #else
    if io.use_scaling {
      VP8LSetError(dec, VP8_STATUS_INVALID_PARAM)
      goto Err
    }
// C: #endif
    if io.use_scaling || WebPIsPremultipliedMode(dec.output.colorspace) {
      // need the alpha-multiply functions for premultiplied output or rescaling
      WebPInitAlphaProcessing()
    }

    if !WebPIsRGBMode(dec.output.colorspace) {
      WebPInitConvertARGBToYUV()
      if dec.output.u.YUVA.a != nil { WebPInitAlphaProcessing() }
    }
    if dec.incremental {
      if (dec.hdr.color_cache_size > 0 &&
          dec.hdr.saved_color_cache.colors == nil) {
        if !VP8LColorCacheInit(&dec.hdr.saved_color_cache, dec.hdr.color_cache.hash_bits) {
          VP8LSetError(dec, VP8_STATUS_OUT_OF_MEMORY)
          goto Err
        }
      }
    }
    dec.state = READ_DATA
  }

  // Decode.
  if !DecodeImageData(dec, dec.pixels, dec.width, dec.height, io.crop_bottom, ProcessRows) {
    goto Err
  }

  params.last_y = dec.last_out_row
  return 1

Err:
  VP8LClear(dec)
  assert.Assert(dec.status != VP8_STATUS_OK)
  return 0
}

//------------------------------------------------------------------------------
