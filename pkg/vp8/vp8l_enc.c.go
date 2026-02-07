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
// main entry for the lossless encoder.
//
// Author: Vikas Arora (vikaas.arora@gmail.com)
//

import (
	"github.com/bufbuild/buf/private/pkg/tmp"
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/libwebp/dsp"
	"github.com/daanv2/go-webp/pkg/libwebp/enc"
	"github.com/daanv2/go-webp/pkg/libwebp/utils"
	"github.com/daanv2/go-webp/pkg/libwebp/webp"
	"github.com/daanv2/go-webp/pkg/stdlib"
	"github.com/daanv2/go-webp/pkg/string"
)

// Maximum number of histogram images (sub-blocks).
const MAX_HUFF_IMAGE_SIZE =2600
const MAX_HUFFMAN_BITS =(MIN_HUFFMAN_BITS + (1 << NUM_HUFFMAN_BITS) - 1)
// Empirical value for which it becomes too computationally expensive to
// compute the best predictor image.
const MAX_PREDICTOR_IMAGE_SIZE =(1 << 14)

// -----------------------------------------------------------------------------
// Palette

// These five modes are evaluated and their respective entropy is computed.
type EntropyIx int

const (
	kDirect EntropyIx = 0
	kSpatial EntropyIx = 1
	kSubGreen EntropyIx = 2
	kSpatialSubGreen EntropyIx = 3
	kPalette EntropyIx = 4
	kPaletteAndSpatial EntropyIx = 5
	kNumEntropyIx EntropyIx = 6
)

type HistoIx int

const (
  kHistoAlpha HistoIx = iota
  kHistoAlphaPred
  kHistoGreen
  kHistoGreenPred
  kHistoRed
  kHistoRedPred
  kHistoBlue
  kHistoBluePred
  kHistoRedSubGreen
  kHistoRedPredSubGreen
  kHistoBlueSubGreen
  kHistoBluePredSubGreen
  kHistoPalette
  kHistoTotal  // Must be last.
)

const NUM_BUCKETS =256

type HistogramBuckets [NUM_BUCKETS]uint32

// Keeping track of histograms, indexed by HistoIx.
// Ideally, this would just be a struct with meaningful fields, but the
// calculation of `entropy_comp` uses the index. One refactoring at a time :)
type Histograms struct {
  category [kHistoTotal]HistogramBuckets;
}

func AddSingleSubGreen(p uint32, r, b HistogramBuckets) {
  green := p >> 8;  // The upper bits are masked away later.
  r[((p >> 16) - green) & 0xff]++
  b[((p >> 0) - green) & 0xff]++
}

func AddSingle(p uint32, a, r, g, b HistogramBuckets) {
  a[(p >> 24) & 0xff]++
  r[(p >> 16) & 0xff]++
  g[(p >> 8) & 0xff]++
  b[(p >> 0) & 0xff]++
}

func HashPix(pix uint32) uint8 {
  // Note that masking with uint(0xffffffff) is for preventing an
  // 'unsigned int overflow' warning. Doesn't impact the compiled code.
  return (((uint64(pix) + (pix >> 19)) * uint64(0x39c5fba7)) & uint64(0xffffffff)) >> 24;
}

func AnalyzeEntropy(/* const */ argb *uint32, width, height, argb_stride, use_palette, palette_size, transform_bits int , /* const */ min_entropy_ix *EntropyIx, /* const */ red_and_blue_always_zero *int) int {
  var histo *Histograms;

  if (use_palette && palette_size <= 16) {
    // In the case of small palettes, we pack 2, 4 or 8 pixels together. In
    // practice, small palettes are better than any other transform.
    *min_entropy_ix = kPalette;
    *red_and_blue_always_zero = 1;
    return 1;
  }

	//   histo = (*Histograms)WebPSafeCalloc(1, sizeof(*histo));
	hist := &Histograms{}

    var i, x, y int
    var prev_row *uint32 = nil;
    var curr_row *uint32 = argb;
    pix_prev := argb[0];  // Skip the first pixel.
    for y = 0; y < height; y++ {
      for x = 0; x < width; x++ {
        pix := curr_row[x];
        pix_diff := VP8LSubPixels(pix, pix_prev);
        pix_prev = pix;
        if ((pix_diff == 0) || (prev_row != nil && pix == prev_row[x])) {
          continue;
        }
        AddSingle(pix, histo.category[kHistoAlpha], histo.category[kHistoRed], histo.category[kHistoGreen], histo.category[kHistoBlue]);
        AddSingle(pix_diff, histo.category[kHistoAlphaPred], histo.category[kHistoRedPred], histo.category[kHistoGreenPred], histo.category[kHistoBluePred]);
        AddSingleSubGreen(pix, histo.category[kHistoRedSubGreen], histo.category[kHistoBlueSubGreen]);
        AddSingleSubGreen(pix_diff, histo.category[kHistoRedPredSubGreen], histo.category[kHistoBluePredSubGreen]);
        {
          // Approximate the palette by the entropy of the multiplicative hash.
          hash := HashPix(pix);
          histo.category[kHistoPalette][hash] = histo.category[kHistoPalette][hash] + 1;
        }
      }
      prev_row = curr_row;
      curr_row += argb_stride;
    }
    {
      var entropy_comp [kHistoTotal]uint64
      var entropy [kNumEntropyIx]uint64
      var k int
      var j int
      last_mode_to_analyze := tenary.If(use_palette, kPalette, kSpatialSubGreen)
      // Let's add one zero to the predicted histograms. The zeros are removed
      // too efficiently by the pix_diff == 0 comparison, at least one of the
      // zeros is likely to exist.
      histo.category[kHistoRedPredSubGreen][0]++
      histo.category[kHistoBluePredSubGreen][0]++
      histo.category[kHistoRedPred][0]++
      histo.category[kHistoGreenPred][0]++
      histo.category[kHistoBluePred][0]++
      histo.category[kHistoAlphaPred][0]++

      for j = 0; j < kHistoTotal; j++ {
        entropy_comp[j] = VP8LBitsEntropy(histo.category[j], NUM_BUCKETS);
      }
      entropy[kDirect] = entropy_comp[kHistoAlpha] + entropy_comp[kHistoRed] +
                         entropy_comp[kHistoGreen] + entropy_comp[kHistoBlue];
      entropy[kSpatial] =
          entropy_comp[kHistoAlphaPred] + entropy_comp[kHistoRedPred] +
          entropy_comp[kHistoGreenPred] + entropy_comp[kHistoBluePred];
      entropy[kSubGreen] =
          entropy_comp[kHistoAlpha] + entropy_comp[kHistoRedSubGreen] +
          entropy_comp[kHistoGreen] + entropy_comp[kHistoBlueSubGreen];
      entropy[kSpatialSubGreen] =
          entropy_comp[kHistoAlphaPred] + entropy_comp[kHistoRedPredSubGreen] +
          entropy_comp[kHistoGreenPred] + entropy_comp[kHistoBluePredSubGreen];
      entropy[kPalette] = entropy_comp[kHistoPalette];

      // When including transforms, there is an overhead in bits from
      // storing them. This overhead is small but matters for small images.
      // For spatial, there are 14 transformations.
      entropy[kSpatial] += (uint64)VP8LSubSampleSize(width, transform_bits) *
                           VP8LSubSampleSize(height, transform_bits) *
                           VP8LFastLog2(14);
      // For color transforms: 24 as only 3 channels are considered in a
      // ColorTransformElement.
      entropy[kSpatialSubGreen] +=
          (uint64)VP8LSubSampleSize(width, transform_bits) *
          VP8LSubSampleSize(height, transform_bits) * VP8LFastLog2(24);
      // For palettes, add the cost of storing the palette.
      // We empirically estimate the cost of a compressed entry as 8 bits.
      // The palette is differential-coded when compressed hence a much
      // lower cost than sizeof(uint32)*8.
      entropy[kPalette] += (palette_size * uint64(8)) << LOG_2_PRECISION_BITS;

      *min_entropy_ix = kDirect;
      for k = kDirect + 1; k <= last_mode_to_analyze; k++ {
        if (entropy[*min_entropy_ix] > entropy[k]) {
          *min_entropy_ix = (EntropyIx)k;
        }
      }
      assert.Assert((int)*min_entropy_ix <= last_mode_to_analyze);
      *red_and_blue_always_zero = 1;
      // Let's check if the histogram of the chosen entropy mode has
      // non-zero red and blue values. If all are zero, we can later skip
      // the cross color optimization.
      {
        kHistoPairs = [5][2]uint8{
            {kHistoRed, kHistoBlue}, {kHistoRedPred, kHistoBluePred}, {kHistoRedSubGreen, kHistoBlueSubGreen}, {kHistoRedPredSubGreen, kHistoBluePredSubGreen}, {kHistoRed, kHistoBlue}}
        const red_histo *HistogramBuckets =
            &histo.category[kHistoPairs[*min_entropy_ix][0]];
        const blue_histo *HistogramBuckets =
            &histo.category[kHistoPairs[*min_entropy_ix][1]];
        for i = 1; i < NUM_BUCKETS; i++ {
          if (((*red_histo)[i] | (*blue_histo)[i]) != 0) {
            *red_and_blue_always_zero = 0;
            break;
          }
        }
      }
    }

    return 1;
}

// Clamp histogram and transform bits.
func ClampBits(width, height int, bits int, min_bits int, max_bits int, image_size_max int) int {
  var image_size int
  bits = (bits < min_bits) ? min_bits : (bits > max_bits) ? max_bits : bits;
  image_size = VP8LSubSampleSize(width, bits) * VP8LSubSampleSize(height, bits);
  while (bits < max_bits && image_size > image_size_max) {
    bits++
    image_size =
        VP8LSubSampleSize(width, bits) * VP8LSubSampleSize(height, bits);
  }
  // In case the bits reduce the image too much, choose the smallest value
  // setting the histogram image size to 1.
  while (bits > min_bits && image_size == 1) {
    image_size = VP8LSubSampleSize(width, bits - 1) *
                 VP8LSubSampleSize(height, bits - 1);
    if image_size != 1 { break }
    --bits;
  }
  return bits;
}

func GetHistoBits(int method, use_palette int, width, height int) int {
  // Make tile size a function of encoding method (Range: 0 to 6).
  histo_bits := (use_palette ? 9 : 7) - method;
  return ClampBits(width, height, histo_bits, MIN_HUFFMAN_BITS, MAX_HUFFMAN_BITS, MAX_HUFF_IMAGE_SIZE);
}

func GetTransformBits(int method, histo_bits int) int {
  max_transform_bits := (method < 4) ? 6 : (method > 4) ? 4 : 5;
  res :=
      (histo_bits > max_transform_bits) ? max_transform_bits : histo_bits;
  assert.Assert(res <= MAX_TRANSFORM_BITS);
  return res;
}

// Set of parameters to be used in each iteration of the cruncher.
const CRUNCH_SUBCONFIGS_MAX =2
type CrunchSubConfig struct {
  lz77 int
  do_no_cache int
}

type CrunchConfig struct {
  entropy_idx int
  palette_sorting_type PaletteSorting
  sub_configs [CRUNCH_SUBCONFIGS_MAX]CrunchSubConfig
  sub_configs_size int
} ;

// +2 because we add a palette sorting configuration for kPalette and
// kPaletteAndSpatial.
const CRUNCH_CONFIGS_MAX =(kNumEntropyIx + 2 * kPaletteSortingNum)

func EncoderAnalyze(/* const */ enc *VP8LEncoder, crunch_configs [CRUNCH_CONFIGS_MAX]CrunchConfig, /*const*/ crunch_configs_size *int, /*const*/ red_and_blue_always_zero *int) int {
  var pic *WebPPicture = enc.pic;
  width := pic.width;
  height := pic.height;
  var config *config.Config = enc.config;
  method := config.Method;
  low_effort := (config.Method == 0);
  var i int
  var use_palette, transform_bits int 
  var n_lz77s int 
  // If set to 0, analyze the cache with the computed cache value. If 1, also
  // analyze with no-cache.
  do_no_cache := 0;
  assert.Assert(pic != nil && pic.argb != nil);

  // Check whether a palette is possible.
  enc.palette_size = GetColorPalette(pic, enc.palette_sorted);
  use_palette = (enc.palette_size <= MAX_PALETTE_SIZE);
  if (!use_palette) {
    enc.palette_size = 0;
  }

  // Empirical bit sizes.
  enc.histo_bits = GetHistoBits(method, use_palette, pic.width, pic.height);
  transform_bits = GetTransformBits(method, enc.histo_bits);
  enc.predictor_transform_bits = transform_bits;
  enc.cross_color_transform_bits = transform_bits;

  if (low_effort) {
    // AnalyzeEntropy is somewhat slow.
    crunch_configs[0].entropy_idx = use_palette ? kPalette : kSpatialSubGreen;
    crunch_configs[0].palette_sorting_type =
        use_palette ? kSortedDefault : kUnusedPalette;
    n_lz77s = 1;
    *crunch_configs_size = 1;
  } else {
     var min_entropy_ix EntropyIx
    // Try out multiple LZ77 on images with few colors.
    n_lz77s = (enc.palette_size > 0 && enc.palette_size <= 16) ? 2 : 1;
    if (!AnalyzeEntropy(pic.argb, width, height, pic.argb_stride, use_palette, enc.palette_size, transform_bits, &min_entropy_ix, red_and_blue_always_zero)) {
      return 0;
    }
    if (method == 6 && config.Quality == 100) {
      do_no_cache = 1;
      // Go brute force on all transforms.
      *crunch_configs_size = 0;
      for i = 0; i < kNumEntropyIx; i++ {
        // We can only apply kPalette or kPaletteAndSpatial if we can indeed use
        // a palette.
        if ((i != kPalette && i != kPaletteAndSpatial) || use_palette) {
          assert.Assert(*crunch_configs_size < CRUNCH_CONFIGS_MAX);
          if (use_palette && (i == kPalette || i == kPaletteAndSpatial)) {
            var sorting_method int
            for (sorting_method = 0; sorting_method < kPaletteSortingNum;
                 ++sorting_method) {
              var typed_sorting_method PaletteSorting =
                  (PaletteSorting)sorting_method;
              // TODO(vrabaud) kSortedDefault should be tested. It is omitted
              // for now for backward compatibility.
              if (typed_sorting_method == kUnusedPalette ||
                  typed_sorting_method == kSortedDefault) {
                continue;
              }
              crunch_configs[(*crunch_configs_size)].entropy_idx = i;
              crunch_configs[(*crunch_configs_size)].palette_sorting_type =
                  typed_sorting_method;
              ++*crunch_configs_size;
            }
          } else {
            crunch_configs[(*crunch_configs_size)].entropy_idx = i;
            crunch_configs[(*crunch_configs_size)].palette_sorting_type =
                kUnusedPalette;
            ++*crunch_configs_size;
          }
        }
      }
    } else {
      // Only choose the guessed best transform.
      *crunch_configs_size = 1;
      crunch_configs[0].entropy_idx = min_entropy_ix;
      crunch_configs[0].palette_sorting_type =
          use_palette ? kMinimizeDelta : kUnusedPalette;
      if (config.Quality >= 75 && method == 5) {
        // Test with and without color cache.
        do_no_cache = 1;
        // If we have a palette, also check in combination with spatial.
        if (min_entropy_ix == kPalette) {
          *crunch_configs_size = 2;
          crunch_configs[1].entropy_idx = kPaletteAndSpatial;
          crunch_configs[1].palette_sorting_type = kMinimizeDelta;
        }
      }
    }
  }
  // Fill in the different LZ77s.
  assert.Assert(n_lz77s <= CRUNCH_SUBCONFIGS_MAX);
  for i = 0; i < *crunch_configs_size; i++ {
    var j int
    for j = 0; j < n_lz77s; j++ {
      assert.Assert(j < CRUNCH_SUBCONFIGS_MAX);
      crunch_configs[i].sub_configs[j].lz77 =
          (j == 0) ? kLZ77Standard | kLZ77RLE : kLZ77Box;
      crunch_configs[i].sub_configs[j].do_no_cache = do_no_cache;
    }
    crunch_configs[i].sub_configs_size = n_lz77s;
  }
  return 1;
}

func EncoderInit(/* const */ enc *VP8LEncoder) int {
  var pic *WebPPicture = enc.pic;
  width := pic.width;
  height := pic.height;
  pix_cnt := width * height;
  // we round the block size up, so we're guaranteed to have
  // at most MAX_REFS_BLOCK_PER_IMAGE blocks used:
  refs_block_size := (pix_cnt - 1) / MAX_REFS_BLOCK_PER_IMAGE + 1;
  var i int
  if !VP8LHashChainInit(&enc.hash_chain, pix_cnt) { return 0  }

  for (i = 0; i < 4; ++i) VP8LBackwardRefsInit(&enc.refs[i], refs_block_size);

  return 1;
}

// Returns false in case of memory error.
func GetHuffBitLengthsAndCodes(/* const */ histogram_image *VP8LHistogramSet, /*const*/ huffman_codes *HuffmanTreeCode) int {
  var i, k int
  ok := 0;
  total_length_size := 0;
  histogram_image_size := histogram_image.size;
  max_num_symbols := 0;
  var buf_rle *uint8 = nil;
  var huff_tree *HuffmanTree = nil;

  // Iterate over all histograms and get the aggregate number of codes used.
  for i = 0; i < histogram_image_size; i++ {
    var histo *VP8LHistogram = histogram_image.histograms[i];
    var codes *HuffmanTreeCode = &huffman_codes[5 * i];
    assert.Assert(histo != nil);
    for k = 0; k < 5; k++ {
      num_symbols :=
          (k == 0)   ? VP8LHistogramNumCodes(histo.palette_code_bits)
          : (k == 4) ? NUM_DISTANCE_CODES
                     : 256;
      codes[k].num_symbols = num_symbols;
      total_length_size += num_symbols;
    }
  }

  // Allocate and Set Huffman codes.
  {
    codes := make([]uint16, total_length_size)
    lengths := make([]uint8, total_length_size)
    // mem_buf = (*uint8)WebPSafeCalloc(total_length_size, sizeof(*lengths) + sizeof(*codes));
    // if mem_buf == nil { goto End }

    // codes = (*uint16)mem_buf;
    // lengths = (*uint8)&codes[total_length_size];
    for i = 0; i < 5 * histogram_image_size; i++ {
      bit_length := huffman_codes[i].num_symbols;
      huffman_codes[i].codes = codes;
      huffman_codes[i].code_lengths = lengths;
      codes += bit_length;
      lengths += bit_length;
      if (max_num_symbols < bit_length) {
        max_num_symbols = bit_length;
      }
    }
  }

//   buf_rle = (*uint8)WebPSafeMalloc(uint64(1), max_num_symbols);
//   huff_tree = (*HuffmanTree)WebPSafeMalloc(uint64(3) * max_num_symbols, sizeof(*huff_tree));
//   if buf_rle == nil || huff_tree == nil { goto End }
  buf_rle := make([]uint8, max_num_symbols)
  huff_tree := make([]HuffmanTree, 3 * max_num_symbols)

  // Create Huffman trees.
  for i = 0; i < histogram_image_size; i++ {
    var codes *HuffmanTreeCode = &huffman_codes[5 * i];
    var histo *VP8LHistogram = histogram_image.histograms[i];
    VP8LCreateHuffmanTree(histo.literal, 15, buf_rle, huff_tree, codes + 0);
    VP8LCreateHuffmanTree(histo.red, 15, buf_rle, huff_tree, codes + 1);
    VP8LCreateHuffmanTree(histo.blue, 15, buf_rle, huff_tree, codes + 2);
    VP8LCreateHuffmanTree(histo.alpha, 15, buf_rle, huff_tree, codes + 3);
    VP8LCreateHuffmanTree(histo.distance, 15, buf_rle, huff_tree, codes + 4);
  }
  ok = 1;
End:
  if (!ok) {
    stdlib.Memset(huffman_codes, 0, 5 * histogram_image_size * sizeof(*huffman_codes));
  }
  return ok;
}

func StoreHuffmanTreeOfHuffmanTreeToBitMask(/* const */ bw *VP8LBitWriter, /* const */ code_length_bitdepth *uint8) {
  // RFC 1951 will calm you down if you are worried about this funny sequence.
  // This sequence is tuned from that, but more weighted for lower symbol count, // and more spiking histograms.
  kStorageOrder = [CODE_LENGTH_CODES]uint8{
      17, 18, 0, 1, 2, 3, 4, 5, 16, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
  var i int
  // Throw away trailing zeros:
  codes_to_store := CODE_LENGTH_CODES;
  for ; codes_to_store > 4; --codes_to_store {
    if (code_length_bitdepth[kStorageOrder[codes_to_store - 1]] != 0) {
      break;
    }
  }
  VP8LPutBits(bw, codes_to_store - 4, 4);
  for i = 0; i < codes_to_store; i++ {
    VP8LPutBits(bw, code_length_bitdepth[kStorageOrder[i]], 3);
  }
}

func ClearHuffmanTreeIfOnlyOneSymbol(/* const */ huffman_code *HuffmanTreeCode) {
  var k int
  count := 0;
  for k = 0; k < huffman_code.num_symbols; k++ {
    if (huffman_code.code_lengths[k] != 0) {
      count++
      if count > 1 { return }
    }
  }
  for k = 0; k < huffman_code.num_symbols; k++ {
    huffman_code.code_lengths[k] = 0;
    huffman_code.codes[k] = 0;
  }
}

func StoreHuffmanTreeToBitMask(/* const */ bw *VP8LBitWriter, /* const */ tokens *HuffmanTreeToken, /* const */ num_tokens int , /* const */ huffman_code *HuffmanTreeCode) {
  var i int
  for i = 0; i < num_tokens; i++ {
    ix := tokens[i].code;
    extra_bits := tokens[i].extra_bits;
    VP8LPutBits(bw, huffman_code.codes[ix], huffman_code.code_lengths[ix]);
    switch (ix) {
      case 16:
        VP8LPutBits(bw, extra_bits, 2);
        break;
      case 17:
        VP8LPutBits(bw, extra_bits, 3);
        break;
      case 18:
        VP8LPutBits(bw, extra_bits, 7);
        break;
    }
  }
}

// 'huff_tree' and 'tokens' are pre-alloacted buffers.
func StoreFullHuffmanCode(/* const */ bw *VP8LBitWriter, /* const */ huff_tree *HuffmanTree, /* const */ tokens *HuffmanTreeToken, /* const */ tree *HuffmanTreeCode) {
  uint8 code_length_bitdepth[CODE_LENGTH_CODES] = {0}
  uint16 code_length_bitdepth_symbols[CODE_LENGTH_CODES] = {0}
  max_tokens := tree.num_symbols;
  var num_tokens int
   var huffman_code HuffmanTreeCode
  huffman_code.num_symbols = CODE_LENGTH_CODES;
  huffman_code.code_lengths = code_length_bitdepth;
  huffman_code.codes = code_length_bitdepth_symbols;

  VP8LPutBits(bw, 0, 1);
  num_tokens = VP8LCreateCompressedHuffmanTree(tree, tokens, max_tokens);
  {
    uint32 histogram[CODE_LENGTH_CODES] = {0}
    uint8 buf_rle[CODE_LENGTH_CODES] = {0}
    var i int
    for i = 0; i < num_tokens; i++ {
      ++histogram[tokens[i].code];
    }

    VP8LCreateHuffmanTree(histogram, 7, buf_rle, huff_tree, &huffman_code);
  }

  StoreHuffmanTreeOfHuffmanTreeToBitMask(bw, code_length_bitdepth);
  ClearHuffmanTreeIfOnlyOneSymbol(&huffman_code);
  {
    trailing_zero_bits := 0;
    trimmed_length := num_tokens;
    var write_trimmed_length int
    var length int
    i := num_tokens;
    while (i-- > 0) {
      ix := tokens[i].code;
      if (ix == 0 || ix == 17 || ix == 18) {
        --trimmed_length;  // discount trailing zeros
        trailing_zero_bits += code_length_bitdepth[ix];
        if (ix == 17) {
          trailing_zero_bits += 3;
        } else if (ix == 18) {
          trailing_zero_bits += 7;
        }
      } else {
        break;
      }
    }
    write_trimmed_length = (trimmed_length > 1 && trailing_zero_bits > 12);
    length = write_trimmed_length ? trimmed_length : num_tokens;
    VP8LPutBits(bw, write_trimmed_length, 1);
    if (write_trimmed_length) {
      if (trimmed_length == 2) {
        VP8LPutBits(bw, 0, 3 + 2);  // nbitpairs=1, trimmed_length=2
      } else {
        nbits := BitsLog2Floor(trimmed_length - 2);
        nbitpairs := nbits / 2 + 1;
        assert.Assert(trimmed_length > 2);
        assert.Assert(nbitpairs - 1 < 8);
        VP8LPutBits(bw, nbitpairs - 1, 3);
        VP8LPutBits(bw, trimmed_length - 2, nbitpairs * 2);
      }
    }
    StoreHuffmanTreeToBitMask(bw, tokens, length, &huffman_code);
  }
}

// 'huff_tree' and 'tokens' are pre-alloacted buffers.
func StoreHuffmanCode(/* const */ bw *VP8LBitWriter, /*const*/ huff_tree *HuffmanTree, /*const*/ tokens *HuffmanTreeToken, /*const*/ huffman_code *HuffmanTreeCode) {
  var i int
  count := 0;
  int symbols[2] = {0, 0}
  kMaxBits := 8;
  kMaxSymbol := 1 << kMaxBits;

  // Check whether it's a small tree.
  for i = 0; i < huffman_code.num_symbols && count < 3; i++ {
    if (huffman_code.code_lengths[i] != 0) {
      if count < 2 { symbols[count] = i }
      count++
    }
  }

  if (count == 0) {  // emit minimal tree for empty cases
    // bits: small tree marker: 1, count-1: 0, large 8-bit code: 0, code: 0
    VP8LPutBits(bw, 0x01, 4);
  } else if (count <= 2 && symbols[0] < kMaxSymbol && symbols[1] < kMaxSymbol) {
    VP8LPutBits(bw, 1, 1);  // Small tree marker to encode 1 or 2 symbols.
    VP8LPutBits(bw, count - 1, 1);
    if (symbols[0] <= 1) {
      VP8LPutBits(bw, 0, 1);  // Code bit for small (1 bit) symbol value.
      VP8LPutBits(bw, symbols[0], 1);
    } else {
      VP8LPutBits(bw, 1, 1);
      VP8LPutBits(bw, symbols[0], 8);
    }
    if (count == 2) {
      VP8LPutBits(bw, symbols[1], 8);
    }
  } else {
    StoreFullHuffmanCode(bw, huff_tree, tokens, huffman_code);
  }
}

func WriteHuffmanCode(/* const */ bw *VP8LBitWriter, /*const*/ code *HuffmanTreeCode, code_index int) {
  depth := code.code_lengths[code_index];
  symbol := code.codes[code_index];
  VP8LPutBits(bw, symbol, depth);
}

static  func WriteHuffmanCodeWithExtraBits(
    const bw *VP8LBitWriter, /*const*/ code *HuffmanTreeCode, code_index int, bits int, n_bits int) {
  depth := code.code_lengths[code_index];
  symbol := code.codes[code_index];
  VP8LPutBits(bw, (bits << depth) | symbol, depth + n_bits);
}

func StoreImageToBitMask(/* const */ bw *VP8LBitWriter, width int, histo_bits int, /*const*/ refs *VP8LBackwardRefs, /*const*/ histogram_symbols *uint32, /*const*/ huffman_codes *HuffmanTreeCode, /*const*/ pic *WebPPicture) int {
  histo_xsize := histo_bits ? VP8LSubSampleSize(width, histo_bits) : 1;
  tile_mask := (histo_bits == 0) ? 0 : -(1 << histo_bits);
  // x and y trace the position in the image.
  x := 0;
  y := 0;
  tile_x := x & tile_mask;
  tile_y := y & tile_mask;
  histogram_ix := (histogram_symbols[0] >> 8) & 0xffff;
  var codes *HuffmanTreeCode = huffman_codes + 5 * histogram_ix;
  VP8LRefsCursor c = VP8LRefsCursorInit(refs);
  while (VP8LRefsCursorOk(&c)) {
    var v *PixOrCopy = c.cur_pos;
    if ((tile_x != (x & tile_mask)) || (tile_y != (y & tile_mask))) {
      tile_x = x & tile_mask;
      tile_y = y & tile_mask;
      histogram_ix = (histogram_symbols[(y >> histo_bits) * histo_xsize +
                                        (x >> histo_bits)] >>
                      8) &
                     0xffff;
      codes = huffman_codes + 5 * histogram_ix;
    }
    if (PixOrCopyIsLiteral(v)) {
      order = []uint8{1, 2, 0, 3}
      for k := 0; k < 4; k++ {
        code := PixOrCopyLiteral(v, order[k]);
        WriteHuffmanCode(bw, codes + k, code);
      }
    } else if (PixOrCopyIsCacheIdx(v)) {
      code := PixOrCopyCacheIdx(v);
      literal_ix := 256 + NUM_LENGTH_CODES + code;
      WriteHuffmanCode(bw, codes, literal_ix);
    } else {
      bits int, n_bits;
      var code int

      distance := PixOrCopyDistance(v);
      VP8LPrefixEncode(v.len, &code, &n_bits, &bits);
      WriteHuffmanCodeWithExtraBits(bw, codes, 256 + code, bits, n_bits);

      // Don't write the distance with the extra bits code since
      // the distance can be up to 18 bits of extra bits, and the prefix
      // 15 bits, totaling to 33, and our PutBits only supports up to 32 bits.
      VP8LPrefixEncode(distance, &code, &n_bits, &bits);
      WriteHuffmanCode(bw, codes + 4, code);
      VP8LPutBits(bw, bits, n_bits);
    }
    x += PixOrCopyLength(v);
    while (x >= width) {
      x -= width;
      y++
    }
    VP8LRefsCursorNext(&c);
  }
  if (bw.error) {
    return WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
  }
  return 1;
}

// Special case of EncodeImageInternal() for cache-bits=0, histo_bits=31.
// pic and percent are for progress.
func EncodeImageNoHuffman(/* const */ bw *VP8LBitWriter, /*const*/ argb *uint32, /*const*/ hash_chain *VP8LHashChain, /*const*/ refs_array *VP8LBackwardRefs, width, height int, quality int, low_effort int, /*const*/ pic *WebPPicture, percent_range int, /*const*/ percent *int) int {
  var i int
  max_tokens := 0;
  var refs *VP8LBackwardRefs;
  var tokens *HuffmanTreeToken = nil;
  var huffman_codes [5]HuffmanTreeCode = {{0, nil, nil}}
  histogram_symbols[1] := {0}  // only one tree, one symbol
  cache_bits := 0;
  var histogram_image *VP8LHistogramSet = nil;
//   var huff_tree *HuffmanTree = (*HuffmanTree)WebPSafeMalloc(
//       uint64(3) * CODE_LENGTH_CODES, sizeof(*huff_tree));
//   if (huff_tree == nil) {
//     WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
//     goto Error;
//   }
  huff_tree := make([]HuffmanTree, 3*CODE_LENGTH_CODES)

  // Calculate backward references from ARGB image.
  if (!VP8LHashChainFill(hash_chain, quality, argb, width, height, low_effort, pic, percent_range / 2, percent)) {
    goto Error;
  }
  if (!VP8LGetBackwardReferences(width, height, argb, quality, /*low_effort=*/0, kLZ77Standard | kLZ77RLE, cache_bits, /*do_no_cache=*/0, hash_chain, refs_array, &cache_bits, pic, percent_range - percent_range / 2, percent)) {
    goto Error;
  }
  refs = &refs_array[0];
  histogram_image = VP8LAllocateHistogramSet(1, cache_bits);
  if (histogram_image == nil) {
    WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
    goto Error;
  }
  VP8LHistogramSetClear(histogram_image);

  // Build histogram image and symbols from backward references.
  VP8LHistogramStoreRefs(refs, /*distance_modifier=*/nil, /*distance_modifier_arg0=*/0, histogram_image.histograms[0]);

  // Create Huffman bit lengths and codes for each histogram image.
  assert.Assert(histogram_image.size == 1);
  if (!GetHuffBitLengthsAndCodes(histogram_image, huffman_codes)) {
    WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
    goto Error;
  }

  // No color cache, no Huffman image.
  VP8LPutBits(bw, 0, 1);

  // Find maximum number of symbols for the huffman tree-set.
  for i = 0; i < 5; i++ {
    var codes *HuffmanTreeCode = &huffman_codes[i];
    if (max_tokens < codes.num_symbols) {
      max_tokens = codes.num_symbols;
    }
  }

//   tokens = (*HuffmanTreeToken)WebPSafeMalloc(max_tokens, sizeof(*tokens));
//   if (tokens == nil) {
//     WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
//     goto Error;
//   }
  tokens := make([]HuffmanTreeToken, max_tokens)

  // Store Huffman codes.
  for i = 0; i < 5; i++ {
    var codes *HuffmanTreeCode = &huffman_codes[i];
    StoreHuffmanCode(bw, huff_tree, tokens, codes);
    ClearHuffmanTreeIfOnlyOneSymbol(codes);
  }

  // Store actual literals.
  if (!StoreImageToBitMask(bw, width, 0, refs, histogram_symbols, huffman_codes, pic)) {
    goto Error;
  }

Error:
  return (pic.error_code == VP8_ENC_OK);
}

// pic and percent are for progress.
func EncodeImageInternal(
    /* const */ bw *VP8LBitWriter, /*const*/ argb *uint32, /*const*/ hash_chain *VP8LHashChain, refs_array [4]VP8LBackwardRefs,
	width, height int, quality int, low_effort int, /*const*/ 
	config *CrunchConfig, cache_bits *int, histogram_bits_in int, 
	init_byte_position uint64 , /*const*/ hdr_size *int, /*const*/ data_size *int,
	 /*const*/ pic *WebPPicture, percent_range int, /*const*/ percent *int) int {

  histogram_image_xysize :=
      VP8LSubSampleSize(width, histogram_bits_in) *
      VP8LSubSampleSize(height, histogram_bits_in);
  remaining_percent := percent_range;
  percent_start := *percent;
  histogram_image *VP8LHistogramSet = nil;
  tmp_histo *VP8LHistogram = nil;
  uint32 i, histogram_image_size = 0;
  bit_array_size := 0;
  tokens *HuffmanTreeToken = nil;
  huffman_codes *HuffmanTreeCode = nil;

//   var huff_tree *HuffmanTree = (*HuffmanTree)WebPSafeMalloc(uint64(3) * CODE_LENGTH_CODES, sizeof(*huff_tree));
//   var histogram_argb *uint32 = (*uint32)WebPSafeMalloc(histogram_image_xysize, sizeof(*histogram_argb));
	huff_tree := make([]HuffmanTree, 3*CODE_LENGTH_CODES)
	histogram_argb := make([]uint32, histogram_image_xysize)

  var sub_configs_idx int
  int cache_bits_init, write_histogram_image;
  VP8LBitWriter bw_init = *bw, bw_best;
  var hdr_size_tmp int
   var hash_chain_histogram VP8LHashChain  // histogram image hash chain
  bw_size_best := ~(uint64)0;
  assert.Assert(histogram_bits_in >= MIN_HUFFMAN_BITS);
  assert.Assert(histogram_bits_in <= MAX_HUFFMAN_BITS);
  assert.Assert(hdr_size != nil);
  assert.Assert(data_size != nil);

  stdlib.Memset(&hash_chain_histogram, 0, sizeof(hash_chain_histogram));
  if (!VP8LBitWriterInit(&bw_best, 0)) {
    WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
    goto Error;
  }

//   // Make sure we can allocate the different objects.
//   if (huff_tree == nil || histogram_argb == nil ||
//       !VP8LHashChainInit(&hash_chain_histogram, histogram_image_xysize)) {
//     WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
//     goto Error;
//   }

  percent_range = remaining_percent / 5;
  if (!VP8LHashChainFill(hash_chain, quality, argb, width, height, low_effort, pic, percent_range, percent)) {
    goto Error;
  }
  percent_start += percent_range;
  remaining_percent -= percent_range;

  // If the value is different from zero, it has been set during the palette
  // analysis.
  cache_bits_init = (*cache_bits == 0) ? MAX_COLOR_CACHE_BITS : *cache_bits;
  // If several iterations will happen, clone into bw_best.
  if ((config.sub_configs_size > 1 || config.sub_configs[0].do_no_cache) &&
      !VP8LBitWriterClone(bw, &bw_best)) {
    WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
    goto Error;
  }

  for (sub_configs_idx = 0; sub_configs_idx < config.sub_configs_size;
       ++sub_configs_idx) {
    const sub_config *CrunchSubConfig =
        &config.sub_configs[sub_configs_idx];
    int cache_bits_best, i_cache;
    i_remaining_percent := remaining_percent / config.sub_configs_size;
    i_percent_range := i_remaining_percent / 4;
    i_remaining_percent -= i_percent_range;

    if (!VP8LGetBackwardReferences(
            width, height, argb, quality, low_effort, sub_config.lz77, cache_bits_init, sub_config.do_no_cache, hash_chain, &refs_array[0], &cache_bits_best, pic, i_percent_range, percent)) {
      goto Error;
    }

    for i_cache = 0; i_cache < (sub_config.do_no_cache ? 2 : 1); i_cache++ {
      cache_bits_tmp := (i_cache == 0) ? cache_bits_best : 0;
      histogram_bits := histogram_bits_in;
      // Speed-up: no need to study the no-cache case if it was already studied
      // in i_cache == 0.
      if i_cache == 1 && cache_bits_best == 0 { break }

      // Reset the bit writer for this iteration.
      VP8LBitWriterReset(&bw_init, bw);

      // Build histogram image and symbols from backward references.
      histogram_image =
          VP8LAllocateHistogramSet(histogram_image_xysize, cache_bits_tmp);
      tmp_histo = VP8LAllocateHistogram(cache_bits_tmp);
      if (histogram_image == nil || tmp_histo == nil) {
        WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
        goto Error;
      }

      i_percent_range = i_remaining_percent / 3;
      i_remaining_percent -= i_percent_range;
      if (!VP8LGetHistoImageSymbols(
              width, height, &refs_array[i_cache], quality, low_effort, histogram_bits, cache_bits_tmp, histogram_image, tmp_histo, histogram_argb, pic, i_percent_range, percent)) {
        goto Error;
      }
      // Create Huffman bit lengths and codes for each histogram image.
      histogram_image_size = histogram_image.size;
      bit_array_size = 5 * histogram_image_size;
    //   huffman_codes = (*HuffmanTreeCode)WebPSafeCalloc(bit_array_size, sizeof(*huffman_codes));
	  huffman_codes = make([]HuffmanTreeCode, bit_array_size)
      // Note: some histogram_image entries may point to tmp_histos[], so the
      // latter need to outlive the following call to
      // GetHuffBitLengthsAndCodes().
      if (!GetHuffBitLengthsAndCodes(histogram_image, huffman_codes)) {
        WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
        goto Error;
      }
      // Free combined histograms.
      histogram_image = nil

      // Free scratch histograms.
      tmp_histo = nil

      // Color Cache parameters.
      if (cache_bits_tmp > 0) {
        VP8LPutBits(bw, 1, 1);
        VP8LPutBits(bw, cache_bits_tmp, 4);
      } else {
        VP8LPutBits(bw, 0, 1);
      }

      // Huffman image + meta huffman.
      histogram_image_size = 0;
      for i = 0; i < histogram_image_xysize; i++ {
        if (histogram_argb[i] >= histogram_image_size) {
          histogram_image_size = histogram_argb[i] + 1;
        }
        histogram_argb[i] <<= 8;
      }

      write_histogram_image = (histogram_image_size > 1);
      VP8LPutBits(bw, write_histogram_image, 1);
      if (write_histogram_image) {
        VP8LOptimizeSampling(histogram_argb, width, height, histogram_bits_in, MAX_HUFFMAN_BITS, &histogram_bits);
        VP8LPutBits(bw, histogram_bits - 2, 3);
        i_percent_range = i_remaining_percent / 2;
        i_remaining_percent -= i_percent_range;
        if (!EncodeImageNoHuffman(
                bw, histogram_argb, &hash_chain_histogram, &refs_array[2], VP8LSubSampleSize(width, histogram_bits), VP8LSubSampleSize(height, histogram_bits), quality, low_effort, pic, i_percent_range, percent)) {
          goto Error;
        }
      }

      // Store Huffman codes.
      {
        max_tokens := 0;
        // Find maximum number of symbols for the huffman tree-set.
        for i = 0; i < 5 * histogram_image_size; i++ {
          var codes *HuffmanTreeCode = &huffman_codes[i];
          if (max_tokens < codes.num_symbols) {
            max_tokens = codes.num_symbols;
          }
        }

        // tokens = (*HuffmanTreeToken)WebPSafeMalloc(max_tokens, sizeof(*tokens));
        // if (tokens == nil) {
        //   WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
        //   goto Error;
        // }
		tokens = make([]HuffmanTreeToken, max_tokens)

        for i = 0; i < 5 * histogram_image_size; i++ {
          var codes *HuffmanTreeCode = &huffman_codes[i];
          StoreHuffmanCode(bw, huff_tree, tokens, codes);
          ClearHuffmanTreeIfOnlyOneSymbol(codes);
        }
      }
      // Store actual literals.
      hdr_size_tmp = (int)(VP8LBitWriterNumBytes(bw) - init_byte_position);
      if (!StoreImageToBitMask(bw, width, histogram_bits, &refs_array[i_cache], histogram_argb, huffman_codes, pic)) {
        goto Error;
      }
      // Keep track of the smallest image so far.
      if (VP8LBitWriterNumBytes(bw) < bw_size_best) {
        bw_size_best = VP8LBitWriterNumBytes(bw);
        *cache_bits = cache_bits_tmp;
        *hdr_size = hdr_size_tmp;
        *data_size =
            (int)(VP8LBitWriterNumBytes(bw) - init_byte_position - *hdr_size);
        VP8LBitWriterSwap(bw, &bw_best);
      }
      tokens = nil;
      if (huffman_codes != nil) {
        huffman_codes = nil;
      }
    }
  }
  VP8LBitWriterSwap(bw, &bw_best);

  if (!WebPReportProgress(pic, percent_start + remaining_percent, percent)) {
    goto Error;
  }

Error:
  VP8LHashChainClear(&hash_chain_histogram);
  return (pic.error_code == VP8_ENC_OK);
}

// -----------------------------------------------------------------------------
// Transforms

func ApplySubtractGreen(/* const */ enc *VP8LEncoder, width, height int, /*const*/ bw *VP8LBitWriter) {
  VP8LPutBits(bw, TRANSFORM_PRESENT, 1);
  VP8LPutBits(bw, SUBTRACT_GREEN_TRANSFORM, 2);
  VP8LSubtractGreenFromBlueAndRed(enc.argb, width * height);
}

func ApplyPredictFilter(/* const */ enc *VP8LEncoder, width, height int, quality int, low_effort int, used_subtract_green int, /*const*/ bw *VP8LBitWriter, percent_range int, /*const*/ percent *int, /*const*/ best_bits *int) int {
  near_lossless_strength :=
      enc.use_palette ? 100 : enc.config.NearLossless;
  max_bits := ClampBits(width, height, enc.predictor_transform_bits, MIN_TRANSFORM_BITS, MAX_TRANSFORM_BITS, MAX_PREDICTOR_IMAGE_SIZE);
  min_bits := ClampBits(
      width, height, max_bits - 2 * (enc.config.Method > 4 ? enc.config.Method - 4 : 0), MIN_TRANSFORM_BITS, MAX_TRANSFORM_BITS, MAX_PREDICTOR_IMAGE_SIZE);

  if (!VP8LResidualImage(width, height, min_bits, max_bits, low_effort, enc.argb, enc.argb_scratch, enc.transform_data, near_lossless_strength, enc.config.Exact, used_subtract_green, enc.pic, percent_range / 2, percent, best_bits)) {
    return 0;
  }
  VP8LPutBits(bw, TRANSFORM_PRESENT, 1);
  VP8LPutBits(bw, PREDICTOR_TRANSFORM, 2);
  assert.Assert(*best_bits >= MIN_TRANSFORM_BITS && *best_bits <= MAX_TRANSFORM_BITS);
  VP8LPutBits(bw, *best_bits - MIN_TRANSFORM_BITS, NUM_TRANSFORM_BITS);
  return EncodeImageNoHuffman(
      bw, enc.transform_data, &enc.hash_chain, &enc.refs[0], VP8LSubSampleSize(width, *best_bits), VP8LSubSampleSize(height, *best_bits), quality, low_effort, enc.pic, percent_range - percent_range / 2, percent);
}

func ApplyCrossColorFilter(/* const */ enc *VP8LEncoder, width, height int, quality int, low_effort int, /*const*/ bw *VP8LBitWriter, percent_range int, /*const*/ percent *int, /*const*/ best_bits *int) int {
  min_bits := enc.cross_color_transform_bits;

  if (!VP8LColorSpaceTransform(width, height, min_bits, quality, enc.argb, enc.transform_data, enc.pic, percent_range / 2, percent, best_bits)) {
    return 0;
  }
  VP8LPutBits(bw, TRANSFORM_PRESENT, 1);
  VP8LPutBits(bw, CROSS_COLOR_TRANSFORM, 2);
  assert.Assert(*best_bits >= MIN_TRANSFORM_BITS && *best_bits <= MAX_TRANSFORM_BITS);
  VP8LPutBits(bw, *best_bits - MIN_TRANSFORM_BITS, NUM_TRANSFORM_BITS);
  return EncodeImageNoHuffman(
      bw, enc.transform_data, &enc.hash_chain, &enc.refs[0], VP8LSubSampleSize(width, *best_bits), VP8LSubSampleSize(height, *best_bits), quality, low_effort, enc.pic, percent_range - percent_range / 2, percent);
}

// -----------------------------------------------------------------------------

func WriteRiffHeader(/* const */ pic *WebPPicture, uint64 riff_size, uint64 vp8l_size) int {
  uint8 riff[RIFF_HEADER_SIZE + CHUNK_HEADER_SIZE + VP8L_SIGNATURE_SIZE] = {
      'R', 'I', 'F', 'F', 0,   0,   0, 0,   'W', 'E', 'B', 'P', 'V', 'P', '8', 'L', 0,   0,   0,   0,   VP8L_MAGIC_BYTE, }
  PutLE32(riff + TAG_SIZE, (uint32)riff_size);
  PutLE32(riff + RIFF_HEADER_SIZE + TAG_SIZE, (uint32)vp8l_size);
  return pic.writer(riff, sizeof(riff), pic);
}

func WriteImageSize(/* const */ pic *WebPPicture, /*const*/ bw *VP8LBitWriter) int {
  width := pic.width - 1;
  height := pic.height - 1;
  assert.Assert(width < WEBP_MAX_DIMENSION && height < WEBP_MAX_DIMENSION);

  VP8LPutBits(bw, width, VP8L_IMAGE_SIZE_BITS);
  VP8LPutBits(bw, height, VP8L_IMAGE_SIZE_BITS);
  return !bw.error;
}

func WriteRealAlphaAndVersion(/* const */ bw *VP8LBitWriter, has_alpha int) int {
  VP8LPutBits(bw, has_alpha, 1);
  VP8LPutBits(bw, VP8L_VERSION, VP8L_VERSION_BITS);
  return !bw.error;
}

func WriteImage(/* const */ pic *WebPPicture, /*const*/ bw *VP8LBitWriter, /*const*/ coded_size *uint64) int {
  var webpll_data *uint8 = VP8LBitWriterFinish(bw);
  webpll_size := VP8LBitWriterNumBytes(bw);
  vp8l_size := VP8L_SIGNATURE_SIZE + webpll_size;
  pad := vp8l_size & 1;
  riff_size := TAG_SIZE + CHUNK_HEADER_SIZE + vp8l_size + pad;
  *coded_size = 0;

  if (bw.error) {
    return WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
  }

  if (!WriteRiffHeader(pic, riff_size, vp8l_size) ||
      !pic.writer(webpll_data, webpll_size, pic)) {
    return WebPEncodingSetError(pic, VP8_ENC_ERROR_BAD_WRITE);
  }

  if (pad) {
    pad_byte[1] := {0}
    if (!pic.writer(pad_byte, 1, pic)) {
      return WebPEncodingSetError(pic, VP8_ENC_ERROR_BAD_WRITE);
    }
  }
  *coded_size = CHUNK_HEADER_SIZE + riff_size;
  return 1;
}

// -----------------------------------------------------------------------------

func ClearTransformBuffer(/* const */ enc *VP8LEncoder) {
  enc.transform_mem = nil;
  enc.transform_mem_size = 0;
}

// Allocates the memory for argb (W x H) buffer, 2 rows of context for
// prediction and transform data.
// Flags influencing the memory allocated:
//  enc.transform_bits
//  enc.use_predict, enc.use_cross_color
func AllocateTransformBuffer(/* const */ enc *VP8LEncoder, width, height int) int {
  image_size := uint64(width * height);
  // VP8LResidualImage needs room for 2 scanlines of uint32 pixels with an extra
  // pixel in each, plus 2 regular scanlines of bytes.
  // TODO(skal): Clean up by using arithmetic in bytes instead of words.
  argb_scratch_size :=
      enc.use_predict ? (width + 1) * 2 + (width * 2 + sizeof(uint32) - 1) /
                                               sizeof(uint32)
                       : 0;
  transform_data_size :=
      (enc.use_predict || enc.use_cross_color)
          ? (uint64)VP8LSubSampleSize(width, MIN_TRANSFORM_BITS) *
                VP8LSubSampleSize(height, MIN_TRANSFORM_BITS)
          : 0;
  max_alignment_in_words := (WEBP_ALIGN_CST + sizeof(uint32) - 1) / sizeof(uint32);
  mem_size := image_size + max_alignment_in_words +
                            argb_scratch_size + max_alignment_in_words +
                            transform_data_size;
  mem *uint32 = enc.transform_mem;
  if (mem == nil || mem_size > enc.transform_mem_size) {
    ClearTransformBuffer(enc);

    // mem = (*uint32)WebPSafeMalloc(mem_size, sizeof(*mem));
    // if (mem == nil) {
    //   return WebPEncodingSetError(enc.pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
    // }
	mem = make([]uint32, mem_size)

	enc.transform_mem = mem;
    enc.transform_mem_size = (uint64)mem_size;
    enc.argb_content = kEncoderNone;
  }
  enc.argb = mem;
  mem = (*uint32)WEBP_ALIGN(mem + image_size);
  enc.argb_scratch = mem;
  mem = (*uint32)WEBP_ALIGN(mem + argb_scratch_size);
  enc.transform_data = mem;

  enc.current_width = width;
  return 1;
}

func MakeInputImageCopy(/* const */ enc *VP8LEncoder) int {
  var picture *WebPPicture = enc.pic;
  width := picture.width;
  height := picture.height;

  if !AllocateTransformBuffer(enc, width, height) { return 0  }
  if enc.argb_content == kEncoderARGB { return 1  }

  {
    dst *uint32 = enc.argb;
    var src *uint32 = picture.argb;
    var y int
    for y = 0; y < height; y++ {
      memcpy(dst, src, width * sizeof(*dst));
      dst += width;
      src += picture.argb_stride;
    }
  }
  enc.argb_content = kEncoderARGB;
  assert.Assert(enc.current_width == width);
  return 1;
}

// -----------------------------------------------------------------------------

const APPLY_PALETTE_GREEDY_MAX =4

func SearchColorGreedy(/* const */ uint32 palette[], palette_size int, color uint32) uint32 {
  (void)palette_size;
  assert.Assert(palette_size < APPLY_PALETTE_GREEDY_MAX);
  assert.Assert(3 == APPLY_PALETTE_GREEDY_MAX - 1);
  if color == palette[0] { return 0  }
  if color == palette[1] { return 1  }
  if color == palette[2] { return 2  }
  return 3;
}

func ApplyPaletteHash0(color uint32) uint32 {
  // Focus on the green color.
  return (color >> 8) & 0xff;
}

const PALETTE_INV_SIZE_BITS =11
const PALETTE_INV_SIZE =(1 << PALETTE_INV_SIZE_BITS)

func ApplyPaletteHash1( color uint32) uint32 {
  // Forget about alpha.
  return (((color & uint(0x00ffffff)) * uint64(4222244071))) >> (32 - PALETTE_INV_SIZE_BITS);
}

func ApplyPaletteHash2(color uint32) uint32 {
  // Forget about alpha.
  return (((color & uint(0x00ffffff)) * ((uint64(1) << 31) - 1))) >> (32 - PALETTE_INV_SIZE_BITS)
}

// Use 1 pixel cache for ARGB pixels.
func APPLY_PALETTE_FOR(COLOR_INDEX int) {
//   do {
    prev_pix := palette[0];
    prev_idx := 0;
    for y = 0; y < height; y++ {
      for x = 0; x < width; x++ {
        pix := src[x];
        if (pix != prev_pix) {
          prev_idx = COLOR_INDEX;
          prev_pix = pix;
        }
        tmp_row[x] = prev_idx;
      }
      VP8LBundleColorMap(tmp_row, width, xbits, dst);
      src += src_stride;
      dst += dst_stride;
    }
//   } while (0)
}

// Remap argb values in src[] to packed palettes entries in dst[]
// using 'row' as a temporary buffer of size 'width'.
// We assume that all src[] values have a corresponding entry in the palette.
// Note: src[] can be the same as dst[]
func ApplyPalette(/* const */ src *uint32,  src_stride uint32, dst *uint32,  dst_stride uint32, /* const */ palette *uint32, palette_size, width, height,  xbits int, /* const */ pic *WebPPicture) int {
  // TODO(skal): this tmp buffer is not needed if VP8LBundleColorMap() can be
  // made to work in-place.
  var x, y int
  
//   var tmp_row *uint8 = (*uint8)WebPSafeMalloc(width, sizeof(*tmp_row));
//   if (tmp_row == nil) {
//     return WebPEncodingSetError(pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
//   }
	tmp_row := make([]uint8, width)

  if (palette_size < APPLY_PALETTE_GREEDY_MAX) {
    APPLY_PALETTE_FOR(SearchColorGreedy(palette, palette_size, pix));
  } else {
    var i, j int
    var buffer [PALETTE_INV_SIZE]uint16
    uint32 (hash_functions *const[])(uint32) = {
        ApplyPaletteHash0, ApplyPaletteHash1, ApplyPaletteHash2}

    // Try to find a perfect hash function able to go from a color to an index
    // within 1 << PALETTE_INV_SIZE_BITS in order to build a hash map to go
    // from color to index in palette.
    for i = 0; i < 3; i++ {
      use_LUT := 1;
      // Set each element in buffer to max uint16.
      stdlib.Memset(buffer, 0xff, sizeof(buffer));
      for j = 0; j < palette_size; j++ {
        ind := hash_functions[i](palette[j]);
        if (buffer[ind] != uint(0xffff)) {
          use_LUT = 0;
          break;
        } else {
          buffer[ind] = j;
        }
      }
      if use_LUT { break }
    }

    if (i == 0) {
      APPLY_PALETTE_FOR(buffer[ApplyPaletteHash0(pix)]);
    } else if (i == 1) {
      APPLY_PALETTE_FOR(buffer[ApplyPaletteHash1(pix)]);
    } else if (i == 2) {
      APPLY_PALETTE_FOR(buffer[ApplyPaletteHash2(pix)]);
    } else {
      uint32 idx_map[MAX_PALETTE_SIZE];
      uint32 palette_sorted[MAX_PALETTE_SIZE];
      PrepareMapToPalette(palette, palette_size, palette_sorted, idx_map);
      APPLY_PALETTE_FOR(
          idx_map[SearchColorNoIdx(palette_sorted, pix, palette_size)]);
    }
  }
  return 1;
}
#undef APPLY_PALETTE_FOR
#undef PALETTE_INV_SIZE_BITS
#undef PALETTE_INV_SIZE
#undef APPLY_PALETTE_GREEDY_MAX

// Note: Expects "enc.palette" to be set properly.
func MapImageFromPalette(/* const */ enc *VP8LEncoder) int {
  var pic *WebPPicture = enc.pic;
  width := pic.width;
  height := pic.height;
  var palette *uint32 = enc.palette;
  palette_size := enc.palette_size;
  var xbits int

  // Replace each input pixel by corresponding palette index.
  // This is done line by line.
  if (palette_size <= 4) {
    xbits = (palette_size <= 2) ? 3 : 2;
  } else {
    xbits = (palette_size <= 16) ? 1 : 0;
  }

  if (!AllocateTransformBuffer(enc, VP8LSubSampleSize(width, xbits), height)) {
    return 0;
  }
  if (!ApplyPalette(pic.argb, pic.argb_stride, enc.argb, enc.current_width, palette, palette_size, width, height, xbits, pic)) {
    return 0;
  }
  enc.argb_content = kEncoderPalette;
  return 1;
}

// Save palette[] to bitstream.
func EncodePalette(/* const */ bw *VP8LBitWriter, low_effort int, /*const*/ enc *VP8LEncoder, percent_range int, /*const*/ percent *int) int {
  var i int
  uint32 tmp_palette[MAX_PALETTE_SIZE];
  palette_size := enc.palette_size;
  var palette *uint32 = enc.palette;
  // If the last element is 0, do not store it and count on automatic palette
  // 0-filling. This can only happen if there is no pixel packing, hence if
  // there are strictly more than 16 colors (after 0 is removed).
  encoded_palette_size :=
      (enc.palette[palette_size - 1] == 0 && palette_size > 17)
          ? palette_size - 1
          : palette_size;
  VP8LPutBits(bw, TRANSFORM_PRESENT, 1);
  VP8LPutBits(bw, COLOR_INDEXING_TRANSFORM, 2);
  assert.Assert(palette_size >= 1 && palette_size <= MAX_PALETTE_SIZE);
  VP8LPutBits(bw, encoded_palette_size - 1, 8);
  for i = encoded_palette_size - 1; i >= 1; --i {
    tmp_palette[i] = VP8LSubPixels(palette[i], palette[i - 1]);
  }
  tmp_palette[0] = palette[0];
  return EncodeImageNoHuffman(bw, tmp_palette, &enc.hash_chain, &enc.refs[0], encoded_palette_size, 1, /*quality=*/20, low_effort, enc.pic, percent_range, percent);
}

// -----------------------------------------------------------------------------
// VP8LEncoder

func VP8LEncoderNew(/* const */ config *config.Config, /*const*/ picture *WebPPicture) *VP8LEncoder {
//   var enc *VP8LEncoder = (*VP8LEncoder)WebPSafeCalloc(uint64(1), sizeof(*enc));
//   if (enc == nil) {
//     WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
//     return nil;
//   }
  enc := &VP8LEncoder{
	config = config;
	pic = picture;
	argb_content = kEncoderNone;
  }

  VP8LEncDspInit();

  return enc;
}

func VP8LEncoderDelete(enc *VP8LEncoder) {
  if (enc != nil) {
    var i int
    VP8LHashChainClear(&enc.hash_chain);
    for (i = 0; i < 4; ++i) {VP8LBackwardRefsClear(&enc.refs[i]);}
    ClearTransformBuffer(enc);
  }
}

// -----------------------------------------------------------------------------
// Main call

type StreamEncodeContext struct {
	config *config.Config
	picture *WebPPicture
	bw *VP8LBitWriter
	enc *VP8LEncoder
	crunch_configs [CRUNCH_CONFIGS_MAX]CrunchConfig
	num_crunch_configs int
	red_and_blue_always_zero int
	stats *WebPAuxStats
} ;

func EncodeStreamHook(input *void, data *void2) int {
  var params *StreamEncodeContext = (*StreamEncodeContext)input;
  var config *config.Config = params.config;
  var picture *WebPPicture = params.picture;
  var bw *VP8LBitWriter = params.bw;
  var enc *VP8LEncoder = params.enc;
  var crunch_configs *CrunchConfig = params.crunch_configs;
  num_crunch_configs := params.num_crunch_configs;
  red_and_blue_always_zero := params.red_and_blue_always_zero;
#if !defined(WEBP_DISABLE_STATS)
  var stats *WebPAuxStats = params.stats;
#endif
  quality := (int)config.Quality;
  low_effort := (config.Method == 0);
#if (WEBP_NEAR_LOSSLESS == 1)
  width := picture.width;
#endif
  height := picture.height;
  byte_position := VP8LBitWriterNumBytes(bw);
  percent := 2;  // for WebPProgressHook
#if (WEBP_NEAR_LOSSLESS == 1)
  use_near_lossless := 0;
#endif
  hdr_size := 0;
  data_size := 0;
  var idx int
  best_size := ~(uint64)0;
  VP8LBitWriter bw_init = *bw, bw_best;
  (void)data2;

  if (!VP8LBitWriterInit(&bw_best, 0) ||
      (num_crunch_configs > 1 && !VP8LBitWriterClone(bw, &bw_best))) {
    WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
    goto Error;
  }

  for idx = 0; idx < num_crunch_configs; idx++ {
    entropy_idx := crunch_configs[idx].entropy_idx;
    remaining_percent := 97 / num_crunch_configs, percent_range;
    predictor_transform_bits := 0, cross_color_transform_bits = 0;
    enc.use_palette =
        (entropy_idx == kPalette) || (entropy_idx == kPaletteAndSpatial);
    enc.use_subtract_green =
        (entropy_idx == kSubGreen) || (entropy_idx == kSpatialSubGreen);
    enc.use_predict = (entropy_idx == kSpatial) ||
                       (entropy_idx == kSpatialSubGreen) ||
                       (entropy_idx == kPaletteAndSpatial);
    // When using a palette, R/B==0, hence no need to test for cross-color.
    if (low_effort || enc.use_palette) {
      enc.use_cross_color = 0;
    } else {
      enc.use_cross_color = red_and_blue_always_zero ? 0 : enc.use_predict;
    }
    // Reset any parameter in the encoder that is set in the previous iteration.
    enc.cache_bits = 0;
    VP8LBackwardRefsClear(&enc.refs[0]);
    VP8LBackwardRefsClear(&enc.refs[1]);

#if (WEBP_NEAR_LOSSLESS == 1)
    // Apply near-lossless preprocessing.
    use_near_lossless =
        (config.NearLossless < 100) && !enc.use_palette && !enc.use_predict;
    if (use_near_lossless) {
      if !AllocateTransformBuffer(enc, width, height) { goto Error }
      if ((enc.argb_content != kEncoderNearLossless) &&
          !VP8ApplyNearLossless(picture, config.NearLossless, enc.argb)) {
        WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
        goto Error;
      }
      enc.argb_content = kEncoderNearLossless;
    } else {
      enc.argb_content = kEncoderNone;
    }
#else
    enc.argb_content = kEncoderNone;
#endif

    // Encode palette
    if (enc.use_palette) {
      if (!PaletteSort(crunch_configs[idx].palette_sorting_type, enc.pic, enc.palette_sorted, enc.palette_size, enc.palette)) {
        WebPEncodingSetError(enc.pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
        goto Error;
      }
      percent_range = remaining_percent / 4;
      if (!EncodePalette(bw, low_effort, enc, percent_range, &percent)) {
        goto Error;
      }
      remaining_percent -= percent_range;
      if !MapImageFromPalette(enc) { goto Error }
      // If using a color cache, do not have it bigger than the number of
      // colors.
      if (enc.palette_size < (1 << MAX_COLOR_CACHE_BITS)) {
        enc.cache_bits = BitsLog2Floor(enc.palette_size) + 1;
      }
    }
    // In case image is not packed.
    if (enc.argb_content != kEncoderNearLossless &&
        enc.argb_content != kEncoderPalette) {
      if !MakeInputImageCopy(enc) { goto Error }
    }

    // -------------------------------------------------------------------------
    // Apply transforms and write transform data.

    if (enc.use_subtract_green) {
      ApplySubtractGreen(enc, enc.current_width, height, bw);
    }

    if (enc.use_predict) {
      percent_range = remaining_percent / 3;
      if (!ApplyPredictFilter(enc, enc.current_width, height, quality, low_effort, enc.use_subtract_green, bw, percent_range, &percent, &predictor_transform_bits)) {
        goto Error;
      }
      remaining_percent -= percent_range;
    }

    if (enc.use_cross_color) {
      percent_range = remaining_percent / 2;
      if (!ApplyCrossColorFilter(enc, enc.current_width, height, quality, low_effort, bw, percent_range, &percent, &cross_color_transform_bits)) {
        goto Error;
      }
      remaining_percent -= percent_range;
    }

    VP8LPutBits(bw, !TRANSFORM_PRESENT, 1);  // No more transforms.

    // -------------------------------------------------------------------------
    // Encode and write the transformed image.
    if (!EncodeImageInternal(
            bw, enc.argb, &enc.hash_chain, enc.refs, enc.current_width, height, quality, low_effort, &crunch_configs[idx], &enc.cache_bits, enc.histo_bits, byte_position, &hdr_size, &data_size, picture, remaining_percent, &percent)) {
      goto Error;
    }

    // If we are better than what we already have.
    if (VP8LBitWriterNumBytes(bw) < best_size) {
      best_size = VP8LBitWriterNumBytes(bw);
      // Store the BitWriter.
      VP8LBitWriterSwap(bw, &bw_best);
#if !defined(WEBP_DISABLE_STATS)
      // Update the stats.
      if (stats != nil) {
        stats.lossless_features = 0;
        if enc.use_predict { stats.lossless_features |= 1 }
        if enc.use_cross_color { stats.lossless_features |= 2 }
        if enc.use_subtract_green { stats.lossless_features |= 4 }
        if enc.use_palette { stats.lossless_features |= 8 }
        stats.histogram_bits = enc.histo_bits;
        stats.transform_bits = predictor_transform_bits;
        stats.cross_color_transform_bits = cross_color_transform_bits;
        stats.cache_bits = enc.cache_bits;
        stats.palette_size = enc.palette_size;
        stats.lossless_size = (int)(best_size - byte_position);
        stats.lossless_hdr_size = hdr_size;
        stats.lossless_data_size = data_size;
      }
#endif
    }
    // Reset the bit writer for the following iteration if any.
    if num_crunch_configs > 1 { VP8LBitWriterReset(&bw_init, bw) }
  }
  VP8LBitWriterSwap(&bw_best, bw);

Error:
  // The hook should return false in case of error.
  return (params.picture.error_code == VP8_ENC_OK);
}

// Encodes the main image stream using the supplied bit writer.
// Returns false in case of error (stored in picture.error_code).
func VP8LEncodeStream(/* const */ config *config.Config, /*const*/ picture *WebPPicture, /*const*/ bw_main *VP8LBitWriter) int {
  var enc_main *VP8LEncoder = VP8LEncoderNew(config, picture);
  enc_side *VP8LEncoder = nil;
  CrunchConfig crunch_configs[CRUNCH_CONFIGS_MAX];
  int num_crunch_configs_main, num_crunch_configs_side = 0;
  var idx int
  red_and_blue_always_zero := 0;
  WebPWorker worker_main, worker_side;
  StreamEncodeContext params_main, params_side;
  // The main thread uses picture.stats, the side thread uses stats_side.
   var stats_side WebPAuxStats
   var bw_side VP8LBitWriter
   var picture_side WebPPicture
  var worker_interface *WebPWorkerInterface = WebPGetWorkerInterface();
  var ok_main int

  if (enc_main == nil || !VP8LBitWriterInit(&bw_side, 0)) {
    VP8LEncoderDelete(enc_main);
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
  }

  // Afunc "garbage value" error from Clang's static analysis tool.
  if (!WebPPictureInit(&picture_side)) {
    goto Error;
  }

  // Analyze image (entropy, num_palettes etc)
  if (!EncoderAnalyze(enc_main, crunch_configs, &num_crunch_configs_main, &red_and_blue_always_zero) ||
      !EncoderInit(enc_main)) {
    WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
    goto Error;
  }

  // Split the configs between the main and side threads (if any).
  if (config.ThreadLevel > 0) {
    num_crunch_configs_side = num_crunch_configs_main / 2;
    for idx = 0; idx < num_crunch_configs_side; idx++ {
      params_side.crunch_configs[idx] =
          crunch_configs[num_crunch_configs_main - num_crunch_configs_side +
                         idx];
    }
    params_side.num_crunch_configs = num_crunch_configs_side;
  }
  num_crunch_configs_main -= num_crunch_configs_side;
  for idx = 0; idx < num_crunch_configs_main; idx++ {
    params_main.crunch_configs[idx] = crunch_configs[idx];
  }
  params_main.num_crunch_configs = num_crunch_configs_main;

  // Fill in the parameters for the thread workers.
  {
    params_size := (num_crunch_configs_side > 0) ? 2 : 1;
    for idx = 0; idx < params_size; idx++ {
      // Create the parameters for each worker.
      var worker *WebPWorker = (idx == 0) ? &worker_main : &worker_side;
      const param *StreamEncodeContext =
          (idx == 0) ? &params_main : &params_side;
      param.config = config;
      param.red_and_blue_always_zero = red_and_blue_always_zero;
      if (idx == 0) {
        param.picture = picture;
        param.stats = picture.stats;
        param.bw = bw_main;
        param.enc = enc_main;
      } else {
        // Create a side picture (error_code is not thread-safe).
        if (!WebPPictureView(picture, /*left=*/0, /*top=*/0, picture.width, picture.height, &picture_side)) {
          assert.Assert(0);
        }
        picture_side.progress_hook = nil;  // Progress hook is not thread-safe.
        param.picture = &picture_side;  // No need to free a view afterwards.
        param.stats = (picture.stats == nil) ? nil : &stats_side;
        // Create a side bit writer.
        if (!VP8LBitWriterClone(bw_main, &bw_side)) {
          WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
          goto Error;
        }
        param.bw = &bw_side;
        // Create a side encoder.
        enc_side = VP8LEncoderNew(config, &picture_side);
        if (enc_side == nil || !EncoderInit(enc_side)) {
          WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
          goto Error;
        }
        // Copy the values that were computed for the main encoder.
        enc_side.histo_bits = enc_main.histo_bits;
        enc_side.predictor_transform_bits = enc_main.predictor_transform_bits;
        enc_side.cross_color_transform_bits =
            enc_main.cross_color_transform_bits;
        enc_side.palette_size = enc_main.palette_size;
        memcpy(enc_side.palette, enc_main.palette, sizeof(enc_main.palette));
        memcpy(enc_side.palette_sorted, enc_main.palette_sorted, sizeof(enc_main.palette_sorted));
        param.enc = enc_side;
      }
      // Create the workers.
      worker_interface.Init(worker);
      worker.data1 = param;
      worker.data2 = nil;
      worker.hook = EncodeStreamHook;
    }
  }

  // Start the second thread if needed.
  if (num_crunch_configs_side != 0) {
    if (!worker_interface.Reset(&worker_side)) {
      WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
      goto Error;
    }

    #if !defined(WEBP_DISABLE_STATS)
    // This line is here and not in the param initialization above to remove a
    // Clang static analyzer warning.
    if (picture.stats != nil) {
      memcpy(&stats_side, picture.stats, sizeof(stats_side));
    }
    #endif
    worker_interface.Launch(&worker_side);
  }
  // Execute the main thread.
  worker_interface.Execute(&worker_main);
  ok_main = worker_interface.Sync(&worker_main);
  worker_interface.End(&worker_main);
  if (num_crunch_configs_side != 0) {
    // Wait for the second thread.
    ok_side := worker_interface.Sync(&worker_side);
    worker_interface.End(&worker_side);
    if (!ok_main || !ok_side) {
      if (picture.error_code == VP8_ENC_OK) {
        assert.Assert(picture_side.error_code != VP8_ENC_OK);
        WebPEncodingSetError(picture, picture_side.error_code);
      }
      goto Error;
    }
    if (VP8LBitWriterNumBytes(&bw_side) < VP8LBitWriterNumBytes(bw_main)) {
      VP8LBitWriterSwap(bw_main, &bw_side);
  #if !defined(WEBP_DISABLE_STATS)
      if (picture.stats != nil) {
        memcpy(picture.stats, &stats_side, sizeof(*picture.stats));
      }
  #endif
    }
  }

Error:
  VP8LEncoderDelete(enc_main);
  VP8LEncoderDelete(enc_side);
  return (picture.error_code == VP8_ENC_OK);
}

#undef CRUNCH_CONFIGS_MAX
#undef CRUNCH_SUBCONFIGS_MAX

// Encodes the picture.
// Returns 0 if config or picture is nil or picture doesn't have valid argb
// input.
func VP8LEncodeImage(/* const */ config *config.Config, /*const*/ picture *WebPPicture) int {
  var  width, height int
  var  has_alpha int
  var coded_size uint64
  percent := 0;
  var initial_size int;
   var bw VP8LBitWriter

  if picture == nil { return 0  }

  if (config == nil || picture.argb == nil) {
    return WebPEncodingSetError(picture, VP8_ENC_ERROR_nil_PARAMETER);
  }

  width = picture.width;
  height = picture.height;
  // Initialize BitWriter with size corresponding to 16 bpp to photo images and
  // 8 bpp for graphical images.
  initial_size = (config.ImageHint == WEBP_HINT_GRAPH) ? width * height
                                                         : width * height * 2;
  if (!VP8LBitWriterInit(&bw, initial_size)) {
    WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
    goto Error;
  }

  if (!WebPReportProgress(picture, 1, &percent)) {
  UserAbort:
    WebPEncodingSetError(picture, VP8_ENC_ERROR_USER_ABORT);
    goto Error;
  }
  // Reset stats (for pure lossless coding)
  if (picture.stats != nil) {
    var stats *WebPAuxStats = picture.stats;
    stdlib.Memset(stats, 0, sizeof(*stats));
    stats.PSNR[0] = 99.f;
    stats.PSNR[1] = 99.f;
    stats.PSNR[2] = 99.f;
    stats.PSNR[3] = 99.f;
    stats.PSNR[4] = 99.f;
  }

  // Write image size.
  if (!WriteImageSize(picture, &bw)) {
    WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
    goto Error;
  }

  has_alpha = WebPPictureHasTransparency(picture);
  // Write the non-trivial Alpha flag and lossless version.
  if (!WriteRealAlphaAndVersion(&bw, has_alpha)) {
    WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
    goto Error;
  }

  if !WebPReportProgress(picture, 2, &percent) { goto UserAbort }

  // Encode main image stream.
  if !VP8LEncodeStream(config, picture, &bw) { goto Error }

  if !WebPReportProgress(picture, 99, &percent) { goto UserAbort }

  // Finish the RIFF chunk.
  if !WriteImage(picture, &bw, &coded_size) { goto Error }

  if !WebPReportProgress(picture, 100, &percent) { goto UserAbort }

#if !defined(WEBP_DISABLE_STATS)
  // Save size.
  if (picture.stats != nil) {
    picture.stats.coded_size += (int)coded_size;
    picture.stats.lossless_size = (int)coded_size;
  }
#endif

  if (picture.extra_info != nil) {
    mb_w := (width + 15) >> 4;
    mb_h := (height + 15) >> 4;
    stdlib.Memset(picture.extra_info, 0, mb_w * mb_h * sizeof(*picture.extra_info));
  }

Error:
  if (bw.error) {
    WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
  }
  return (picture.error_code == VP8_ENC_OK);
}

//------------------------------------------------------------------------------
