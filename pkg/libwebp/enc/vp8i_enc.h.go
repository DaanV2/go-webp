package enc

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
//   WebP encoder: internal header.
//
// Author: Skal (pascal.massimino@gmail.com)


import "github.com/daanv2/go-webp/pkg/string"  // for memcpy()

import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

#ifdef __cplusplus
extern "C" {
#endif

//------------------------------------------------------------------------------
// Various defines and enums

// version numbers
const ENC_MAJ_VERSION =1
const ENC_MIN_VERSION =6
const ENC_REV_VERSION =0

enum {
  MAX_LF_LEVELS = 64,       // Maximum loop filter level
  MAX_VARIABLE_LEVEL = 67,  // last (inclusive) level with variable cost
  MAX_LEVEL = 2047          // max level (note: max codable is 2047 + 67)
}

type VP8RDLevel int

const (            // Rate-distortion optimization levels
  RD_OPT_NONE VP8RDLevel = 0,        // no rd-opt
  RD_OPT_BASIC VP8RDLevel = 1,       // basic scoring (no trellis)
  RD_OPT_TRELLIS VP8RDLevel = 2,     // perform trellis-quant on the final decision only
  RD_OPT_TRELLIS_ALL VP8RDLevel = 3  // trellis-quant for every scoring (much slower)
)

// YUV-cache parameters. Cache is 32-bytes wide (= one cacheline).
// The original or reconstructed samples can be accessed using VP8Scan[].
// The predicted blocks can be accessed using offsets to 'yuv_p' and
// the arrays *VP8ModeOffsets[].
// * YUV Samples area ('yuv_in'/'yuv_out'/'yuv_out2')
//   (see VP8Scan[] for accessing the blocks, along with
//   Y_OFF_ENC/U_OFF_ENC/V_OFF_ENC):
//             +----+----+
//  Y_OFF_ENC  |YYYY|UUVV|
//  U_OFF_ENC  |YYYY|UUVV|
//  V_OFF_ENC  |YYYY|....| <- 25% wasted U/V area
//             |YYYY|....|
//             +----+----+
// * Prediction area ('yuv_p', size = PRED_SIZE_ENC)
//   Intra16 predictions (16x16 block each, two per row):
//         |I16DC16|I16TM16|
//         |I16VE16|I16HE16|
//   Chroma U/V predictions (16x8 block each, two per row):
//         |C8DC8|C8TM8|
//         |C8VE8|C8HE8|
//   Intra 4x4 predictions (4x4 block each)
//         |I4DC4 I4TM4 I4VE4 I4HE4|I4RD4 I4VR4 I4LD4 I4VL4|
//         |I4HD4 I4HU4 I4TMP .....|.......................| <- ~31% wasted
const YUV_SIZE_ENC =(BPS * 16)
const PRED_SIZE_ENC =(32 * BPS + 16 * BPS + 8 * BPS)  // I16+Chroma+I4 preds
const Y_OFF_ENC =(0)
const U_OFF_ENC =(16)
const V_OFF_ENC =(16 + 8)

extern const uint16 VP8Scan[16];
extern const uint16 VP8UVModeOffsets[4];
extern const uint16 VP8I16ModeOffsets[4];

// Layout of prediction blocks
// intra 16x16
const I16DC16 = (0 * 16 * BPS)
const I16TM16 = (I16DC16 + 16)
const I16VE16 = (1 * 16 * BPS)
const I16HE16 = (I16VE16 + 16)
// chroma 8x8, two U/V blocks side by side (hence: 16x8 each)
const C8DC8 = (2 * 16 * BPS)
const C8TM8 = (C8DC8 + 1 * 16)
const C8VE8 = (2 * 16 * BPS + 8 * BPS)
const C8HE8 = (C8VE8 + 1 * 16)
// intra 4x4
const I4DC4 = (3 * 16 * BPS + 0)
const I4TM4 = (I4DC4 + 4)
const I4VE4 = (I4DC4 + 8)
const I4HE4 = (I4DC4 + 12)
const I4RD4 = (I4DC4 + 16)
const I4VR4 = (I4DC4 + 20)
const I4LD4 = (I4DC4 + 24)
const I4VL4 = (I4DC4 + 28)
const I4HD4 = (3 * 16 * BPS + 4 * BPS)
const I4HU4 = (I4HD4 + 4)
const I4TMP = (I4HD4 + 8)

typedef int64 score_t;  // type used for scores, rate, distortion
// Note that MAX_COST is not the maximum allowed by sizeof(score_t),
// in order to allow overflowing computations.
const MAX_COST =((score_t)0x7fffffffffffffLL)

const QFIX = 17
#define BIAS(b) ((b) << (QFIX - 8))
// Fun fact: this is the _only_ line where we're actually being lossy and
// discarding bits.
static  int QUANTDIV(uint32 n, uint32 iQ, uint32 B) {
  return (int)((n * iQ + B) >> QFIX);
}

// Uncomment the following to remove token-buffer code:
// #define DISABLE_TOKEN_BUFFER

// quality below which error-diffusion is enabled
const ERROR_DIFFUSION_QUALITY =98

//------------------------------------------------------------------------------
// Headers

type proba_t uint32  // 16b + 16b
type ProbaArray [NUM_CTX][NUM_PROBAS]uint8
type StatsArray [NUM_CTX][NUM_PROBAS]proba_t
type CostArray [NUM_CTX][MAX_VARIABLE_LEVEL + 1]uint16
type CostArrayPtr [NUM_CTX]*uint16  // for easy casting
type CostArrayMap [16][NUM_CTX]*uint16;
type LFStats [NUM_MB_SEGMENTS][MAX_LF_LEVELS]float64  // filter stats

// segment features
type VP8EncSegmentHeader struct {
	num_segments int  // Actual number of segments. 1 segment only = unused.
	// whether to update the segment map or not.
	// must be 0 if there's only 1 segment.
	update_map int    
	size int          // bit-cost for transmitting the segment map
}

// Struct collecting all frame-persistent probabilities.
type VP8EncProba struct {
  segments [3]uint8;  // probabilities for segment tree
  skip_proba uint8;   // final probability of being skipped.
  coeffs [NUM_TYPES][NUM_BANDS]ProbaArray;     // 1056 bytes
  stats [NUM_TYPES][NUM_BANDS]StatsArray;      // 4224 bytes
  level_cost [NUM_TYPES][NUM_BANDS]CostArray;  // 13056 bytes
  remapped_costs [NUM_TYPES]CostArrayMap;      // 1536 bytes
  dirty int            // if true, need to call VP8CalculateLevelCosts()
  use_skip_proba int   // Note: we always use skip_proba for now.
  nb_skip int          // number of skipped blocks
}

// Filter parameters. Not actually used in the code (we don't perform
// the in-loop filtering), but filled from user's config
type VP8EncFilterHeader struct {
   simple int         // filtering type: 0=complex, 1=simple
   level int          // base filter level [0..63]
   sharpness int      // [0..7]
   i4x4_lf_delta int  // delta filter level for i4x4 relative to i16x16
}

//------------------------------------------------------------------------------
// Informations about the macroblocks.

type VP8MBInfo struct {
  // block type
  unsigned int type : 2;  // 0=i4x4, 1=i16x16
  unsigned int uv_mode : 2;
  unsigned int skip : 1;
  unsigned int segment : 2;
  uint8 alpha;  // quantization-susceptibility
} ;

type VP8Matrix struct {
  uint16 q[16];        // quantizer steps
  uint16 iq[16];       // reciprocals, fixed point.
  uint32 bias[16];     // rounding bias
  uint32 zthresh[16];  // value below which a coefficient is zeroed
  uint16 sharpen[16];  // frequency boosters for slight sharpening
} ;

type VP8SegmentInfo struct {
  VP8Matrix y1, y2, uv;  // quantization matrices
  int alpha;      // quant-susceptibility, range [-127,127]. Zero is neutral.
                  // Lower values indicate a lower risk of blurriness.
  int beta;       // filter-susceptibility, range [0,255].
  int quant;      // final segment quantizer.
  int fstrength;  // final in-loop filtering strength
  int max_edge;   // max edge delta (for filtering strength)
  int min_disto;  // minimum distortion required to trigger filtering record
  // reactivities
  int lambda_i16, lambda_i4, lambda_uv;
  int lambda_mode, lambda_trellis, tlambda;
  int lambda_trellis_i16, lambda_trellis_i4, lambda_trellis_uv;

  // lambda values for distortion-based evaluation
  score_t i4_penalty;  // penalty for using Intra4
} ;

type DError [2 /* u/v */][2 /* top or left */]int8

// Handy transient struct to accumulate score and info during RD-optimization
// and mode evaluation.
type VP8ModeScore struct {
  score_t D, SD;            // Distortion, spectral distortion
  score_t H, R, score;      // header bits, rate, score.
  int16 y_dc_levels[16];  // Quantized levels for luma-DC, luma-AC, chroma.
  int16 y_ac_levels[16][16];
  int16 uv_levels[4 + 4][16];
  int mode_i16;          // mode number for intra16 prediction
  uint8 modes_i4[16];  // mode numbers for intra4 predictions
  int mode_uv;           // mode number of chroma prediction
  uint32 nz;           // non-zero blocks
  int8 derr[2][3];     // DC diffusion errors for U/V for blocks #1/2/3
} ;

// Iterator structure to iterate through macroblocks, pointing to the
// right neighbouring data (samples, predictions, contexts, ...)
type VP8EncIterator struct {
  int x, y;           // current macroblock
  yuv_in *uint8;    // input samples
  yuv_out *uint8;   // output samples
  yuv_out *uint82;  // secondary buffer swapped with yuv_out.
  yuv_p *uint8;     // scratch buffer for prediction
  enc *VP8Encoder;    // back-pointer
  mb *VP8MBInfo;      // current macroblock
  bw *VP8BitWriter;   // current bit-writer
  preds *uint8;     // intra mode predictors (4x4 blocks)
  nz *uint32;       // non-zero pattern
#if WEBP_AARCH64 && BPS == 32
  uint8 i4_boundary[40];  // 32+8 boundary samples needed by intra4x4
#else
  uint8 i4_boundary[37];  // 32+5 boundary samples needed by intra4x4
#endif
  i *uint84_top;           // pointer to the current top boundary sample
  int i4;                    // current intra4x4 mode being tested
  int top_nz[9];             // top-non-zero context.
  int left_nz[9];            // left-non-zero. left_nz[8] is independent.
  uint64 bit_count[4][3];  // bit counters for coded levels.
  uint64 luma_bits;        // macroblock bit-cost for luma
  uint64 uv_bits;          // macroblock bit-cost for chroma
  lf_stats *LFStats;         // filter stats (borrowed from enc)
  int do_trellis;            // if true, perform extra level optimisation
  int count_down;            // number of mb still to be processed
  int count_down0;           // starting counter value (for progress)
  int percent0;              // saved initial progress percent

  DError left_derr;  // left error diffusion (u/v)
  top_derr *DError;  // top diffusion error - nil if disabled

  y_left *uint8;  // left luma samples (addressable from index -1 to 15).
  u_left *uint8;  // left u samples (addressable from index -1 to 7)
  v_left *uint8;  // left v samples (addressable from index -1 to 7)

  y_top *uint8;   // top luma samples at position 'x'
  uv_top *uint8;  // top u/v samples at position 'x', packed as 16 bytes

  // memory for storing y/u/v_left
  uint8 yuv_left_mem[17 + 16 + 16 + 8 + WEBP_ALIGN_CST];
  // memory for *yuv
  uint8 yuv_mem[3 * YUV_SIZE_ENC + PRED_SIZE_ENC + WEBP_ALIGN_CST];
} ;

// in iterator.c
// must be called first
func VP8IteratorInit(const enc *VP8Encoder, const it *VP8EncIterator);
// reset iterator position to row 'y'
func VP8IteratorSetRow(const it *VP8EncIterator, int y);
// set count down (=number of iterations to go)
func VP8IteratorSetCountDown(const it *VP8EncIterator, int count_down);
// return true if iteration is finished
int VP8IteratorIsDone(const it *VP8EncIterator);
// Import uncompressed samples from source.
// If tmp_32 is not nil, import boundary samples too.
// tmp_32 is a 32-bytes scratch buffer that must be aligned in memory.
func VP8IteratorImport(const it *VP8EncIterator, const tmp_ *uint832);
// export decimated samples
func VP8IteratorExport(const it *VP8EncIterator);
// go to next macroblock. Returns false if not finished.
int VP8IteratorNext(const it *VP8EncIterator);
// save the 'yuv_out' boundary values to 'top'/'left' arrays for next
// iterations.
func VP8IteratorSaveBoundary(const it *VP8EncIterator);
// Report progression based on macroblock rows. Return 0 for user-abort request.
int VP8IteratorProgress(const it *VP8EncIterator, int delta);
// Intra4x4 iterations
func VP8IteratorStartI4(const it *VP8EncIterator);
// returns true if not done.
int VP8IteratorRotateI4(const it *VP8EncIterator, const yuv_out *uint8);

// Non-zero context setup/teardown
func VP8IteratorNzToBytes(const it *VP8EncIterator);
func VP8IteratorBytesToNz(const it *VP8EncIterator);

// Helper functions to set mode properties
func VP8SetIntra16Mode(const it *VP8EncIterator, int mode);
func VP8SetIntra4Mode(const it *VP8EncIterator, const modes *uint8);
func VP8SetIntraUVMode(const it *VP8EncIterator, int mode);
func VP8SetSkip(const it *VP8EncIterator, int skip);
func VP8SetSegment(const it *VP8EncIterator, int segment);

//------------------------------------------------------------------------------
// Paginated token buffer

typedef struct VP8Tokens VP8Tokens;  // struct details in token.c

type VP8TBuffer struct {
#if !defined(DISABLE_TOKEN_BUFFER)
  pages *VP8Tokens;       // first page
  *VP8Tokens* last_page;  // last page
  tokens *uint16;       // set to (*last_page).tokens
  int left;               // how many free tokens left before the page is full
  int page_size;          // number of tokens per page
#endif
  int error;  // true in case of malloc error
} ;

// initialize an empty buffer
func VP8TBufferInit(const b *VP8TBuffer, int page_size);
func VP8TBufferClear(const b *VP8TBuffer);  // de-allocate pages memory

#if !defined(DISABLE_TOKEN_BUFFER)

// Finalizes bitstream when probabilities are known.
// Deletes the allocated token memory if final_pass is true.
int VP8EmitTokens(const b *VP8TBuffer, const bw *VP8BitWriter, const probas *uint8, int final_pass);

// record the coding of coefficients without knowing the probabilities yet
int VP8RecordCoeffTokens(int ctx, const struct const res *VP8Residual, const tokens *VP8TBuffer);

// Estimate the final coded size given a set of 'probas'.
uint64 VP8EstimateTokenSize(const b *VP8TBuffer, const probas *uint8);

#endif  // !DISABLE_TOKEN_BUFFER

//------------------------------------------------------------------------------
// VP8Encoder

type VP8Encoder struct {
  const config *WebPConfig;  // user configuration and parameters
  pic *WebPPicture;          // input / output picture

  // headers
  VP8EncFilterHeader filter_hdr;    // filtering information
  VP8EncSegmentHeader segment_hdr;  // segment information

  int profile;  // VP8's profile, deduced from Config.

  // dimension, in macroblock units.
  int mb_w, mb_h;
  int preds_w;  // stride of the prediction plane *preds (=4*mb_w + 1)

  // number of partitions (1, 2, 4 or 8 = MAX_NUM_PARTITIONS)
  int num_parts;

  // per-partition boolean decoders.
  VP8BitWriter bw;                         // part0
  VP8BitWriter parts[MAX_NUM_PARTITIONS];  // token partitions
  VP8TBuffer tokens;                       // token buffer

  int percent;  // for progress

  // transparency blob
  int has_alpha;
  alpha_data *uint8;  // non-nil if transparency is present
  uint32 alpha_data_size;
  WebPWorker alpha_worker;

  // quantization info (one set of DC/AC dequant factor per segment)
  VP8SegmentInfo dqm[NUM_MB_SEGMENTS];
  int base_quant;  // nominal quantizer value. Only used
                   // for relative coding of segments' quant.
  int alpha;       // global susceptibility (<=> complexity)
  int uv_alpha;    // U/V quantization susceptibility
  // global offset of quantizers, shared by all segments
  int dq_y1_dc;
  int dq_y2_dc, dq_y2_ac;
  int dq_uv_dc, dq_uv_ac;

  // probabilities and statistics
  VP8EncProba proba;
  uint64 sse[4];     // sum of Y/U/V/A squared errors for all macroblocks
  uint64 sse_count;  // pixel count for the sse[] stats
  int coded_size;
  int residual_bytes[3][4];
  int block_count[3];

  // quality/speed settings
  int method;               // 0=fastest, 6=best/slowest.
  VP8RDLevel rd_opt_level;  // Deduced from method.
  int max_i4_header_bits;   // partition #0 safeness factor
  int mb_header_limit;      // rough limit for header bits per MB
  int thread_level;         // derived from config.thread_level
  int do_search;            // derived from config.target_XXX
  int use_tokens;           // if true, use token buffer

  // Memory
  mb_info *VP8MBInfo;  // contextual macroblock infos (mb_w + 1)
  preds *uint8;      // predictions modes: (4*mb_w+1) * (4*mb_h+1)
  nz *uint32;        // non-zero bit context: mb_w+1
  y_top *uint8;      // top luma samples.
  uv_top *uint8;     // top u/v samples.
                       // U and V are packed into 16 bytes (8 U + 8 V)
  lf_stats *LFStats;   // autofilter stats (if nil, autofilter is off)
  top_derr *DError;    // diffusion error (nil if disabled)
}

//------------------------------------------------------------------------------
// internal functions. Not public.

// in tree.c
extern const uint8 VP8CoeffsProba0[NUM_TYPES][NUM_BANDS][NUM_CTX][NUM_PROBAS];
extern const uint8 VP8CoeffsUpdateProba[NUM_TYPES][NUM_BANDS][NUM_CTX]
                                         [NUM_PROBAS];
// Reset the token probabilities to their initial (default) values
func VP8DefaultProbas(const enc *VP8Encoder);
// Write the token probabilities
func VP8WriteProbas(const bw *VP8BitWriter, const probas *VP8EncProba);
// Writes the partition #0 modes (that is: all intra modes)
func VP8CodeIntraModes(const enc *VP8Encoder);

// in syntax.c
// Generates the final bitstream by coding the partition0 and headers,
// and appending an assembly of all the pre-coded token partitions.
// Return true if everything is ok.
int VP8EncWrite(const enc *VP8Encoder);
// Release memory allocated for bit-writing in VP8EncLoop & seq.
func VP8EncFreeBitWriters(const enc *VP8Encoder);

// in frame.c
extern const uint8 VP8Cat3[];
extern const uint8 VP8Cat4[];
extern const uint8 VP8Cat5[];
extern const uint8 VP8Cat6[];

// Form all the four Intra16x16 predictions in the 'yuv_p' cache
func VP8MakeLuma16Preds(const it *VP8EncIterator);
// Form all the four Chroma8x8 predictions in the 'yuv_p' cache
func VP8MakeChroma8Preds(const it *VP8EncIterator);
// Rate calculation
int VP8GetCostLuma16(const it *VP8EncIterator, const rd *VP8ModeScore);
int VP8GetCostLuma4(const it *VP8EncIterator, const int16 levels[16]);
int VP8GetCostUV(const it *VP8EncIterator, const rd *VP8ModeScore);
// Main coding calls
int VP8EncLoop(const enc *VP8Encoder);
int VP8EncTokenLoop(const enc *VP8Encoder);

// in webpenc.c
// Assign an error code to a picture. Return false for convenience.
int WebPEncodingSetError(const pic *WebPPicture, WebPEncodingError error);
int WebPReportProgress(const pic *WebPPicture, int percent, const percent_store *int);

// in analysis.c
// Main analysis loop. Decides the segmentations and complexity.
// Assigns a first guess for Intra16 and 'uvmode' prediction modes.
int VP8EncAnalyze(const enc *VP8Encoder);

// in quant.c
// Sets up segment's quantization values, 'base_quant' and filter strengths.
func VP8SetSegmentParams(const enc *VP8Encoder, float quality);
// Pick best modes and fills the levels. Returns true if skipped.
int VP8Decimate(WEBP_RESTRICT const it *VP8EncIterator, WEBP_RESTRICT const rd *VP8ModeScore, VP8RDLevel rd_opt);

// in alpha.c
func VP8EncInitAlpha(const enc *VP8Encoder);   // initialize alpha compression
int VP8EncStartAlpha(const enc *VP8Encoder);   // start alpha coding process
int VP8EncFinishAlpha(const enc *VP8Encoder);  // finalize compressed data
int VP8EncDeleteAlpha(const enc *VP8Encoder);  // delete compressed data

// autofilter
func VP8InitFilter(const it *VP8EncIterator);
func VP8StoreFilterStats(const it *VP8EncIterator);
func VP8AdjustFilterStrength(const it *VP8EncIterator);

// returns the approximate filtering strength needed to smooth a edge
// step of 'delta', given a sharpness parameter 'sharpness'.
int VP8FilterStrengthFromDelta(int sharpness, int delta);

// misc utils for picture_*.c:

// Returns true if 'picture' is non-nil and dimensions/colorspace are within
// their valid ranges. If returning false, the 'error_code' in 'picture' is
// updated.
int WebPValidatePicture(const picture *WebPPicture);

// Remove reference to the ARGB/YUVA buffer (doesn't free anything).
func WebPPictureResetBuffers(const picture *WebPPicture);

// Allocates ARGB buffer according to set width/height (previous one is
// always free'd). Preserves the YUV(A) buffer. Returns false in case of error
// (invalid param, out-of-memory).
int WebPPictureAllocARGB(const picture *WebPPicture);

// Allocates YUVA buffer according to set width/height (previous one is always
// free'd). Uses picture.csp to determine whether an alpha buffer is needed.
// Preserves the ARGB buffer.
// Returns false in case of error (invalid param, out-of-memory).
int WebPPictureAllocYUVA(const picture *WebPPicture);

// Replace samples that are fully transparent by 'color' to help compressibility
// (no guarantee, though). Assumes pic.use_argb is true.
func WebPReplaceTransparentPixels(const pic *WebPPicture, uint32 color);

//------------------------------------------------------------------------------

#ifdef __cplusplus
}  // extern "C"
#endif

#endif  // WEBP_ENC_VP8I_ENC_H_
