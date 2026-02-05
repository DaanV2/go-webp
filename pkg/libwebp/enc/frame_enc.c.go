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
//   frame coding and analysis
//
// Author: Skal (pascal.massimino@gmail.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/math"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/decoder"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"  // RIFF constants
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

const SEGMENT_VISU =0
const DEBUG_SEARCH =0  // useful to track search convergence

//------------------------------------------------------------------------------
// multi-pass convergence

const HEADER_SIZE_ESTIMATE = (RIFF_HEADER_SIZE + CHUNK_HEADER_SIZE + VP8_FRAME_HEADER_SIZE)
const DQ_LIMIT =0.4  // convergence is considered reached if dq < DQ_LIMIT
// we allow 2k of extra head-room in PARTITION0 limit.
const PARTITION0_SIZE_LIMIT =((VP8_MAX_PARTITION0_SIZE - uint64(2048)) << 11)

func Clamp(float v, float min, float max) float {
  return (v < min) ? min : (v > max) ? max : v;
}

type PassStats struct {  // struct for organizing convergence in either size or PSNR
  var is_first int
  float dq;
  float q, last_q;
  float qmin, qmax;
  double value, last_value;  // PSNR or size
  double target;
  var do_size_search int
} ;

func InitPassStats(/* const */ enc *VP8Encoder, /*const*/ s *PassStats) int {
  target_size := (uint64)enc.config.target_size;
  do_size_search := (target_size != 0);
  const float target_PSNR = enc.config.target_PSNR;

  s.is_first = 1;
  s.dq = 10.f;
  s.qmin = 1.f * enc.config.qmin;
  s.qmax = 1.f * enc.config.qmax;
  s.q = s.last_q = Clamp(enc.config.quality, s.qmin, s.qmax);
  s.target = do_size_search       ? (double)target_size
              : (target_PSNR > 0.) ? target_PSNR
                                   : 40.;  // default, just in case
  s.value = s.last_value = 0.;
  s.do_size_search = do_size_search;
  return do_size_search;
}

func ComputeNextQ(/* const */ s *PassStats) float {
  float dq;
  if (s.is_first) {
    dq = (s.value > s.target) ? -s.dq : s.dq;
    s.is_first = 0;
  } else if (s.value != s.last_value) {
    const double slope = (s.target - s.value) / (s.last_value - s.value);
    dq = (float)(slope * (s.last_q - s.q));
  } else {
    dq = 0.;  // we're done?!
  }
  // Limit variable to afunc large swings.
  s.dq = Clamp(dq, -30.f, 30.f);
  s.last_q = s.q;
  s.last_value = s.value;
  s.q = Clamp(s.q + s.dq, s.qmin, s.qmax);
  return s.q;
}

//------------------------------------------------------------------------------
// Tables for level coding

const uint8 VP8Cat3[] = {173, 148, 140}
const uint8 VP8Cat4[] = {176, 155, 140, 135}
const uint8 VP8Cat5[] = {180, 157, 141, 134, 130}
const uint8 VP8Cat6[] = {254, 254, 243, 230, 196, 177, 153, 140, 133, 130, 129}

//------------------------------------------------------------------------------
// Reset the statistics about: number of skips, token proba, level cost,...

func ResetStats(/* const */ enc *VP8Encoder) {
  var proba *VP8EncProba = &enc.proba;
  VP8CalculateLevelCosts(proba);
  proba.nb_skip = 0;
}

//------------------------------------------------------------------------------
// Skip decision probability

const SKIP_PROBA_THRESHOLD =250  // value below which using skip_proba is OK.

func CalcSkipProba(uint64 nb, uint64 total) int {
  return (int)(total ? (total - nb) * 255 / total : 255);
}

// Returns the bit-cost for coding the skip probability.
func FinalizeSkipProba(/* const */ enc *VP8Encoder) int {
  var proba *VP8EncProba = &enc.proba;
  nb_mbs := enc.mb_w * enc.mb_h;
  nb_events := proba.nb_skip;
  var size int
  proba.skip_proba = CalcSkipProba(nb_events, nb_mbs);
  proba.use_skip_proba = (proba.skip_proba < SKIP_PROBA_THRESHOLD);
  size = 256;  // 'use_skip_proba' bit
  if (proba.use_skip_proba) {
    size += nb_events * VP8BitCost(1, proba.skip_proba) +
            (nb_mbs - nb_events) * VP8BitCost(0, proba.skip_proba);
    size += 8 * 256;  // cost of signaling the 'skip_proba' itself.
  }
  return size;
}

// Collect statistics and deduce probabilities for next coding pass.
// Return the total bit-cost for coding the probability updates.
func CalcTokenProba(int nb, int total) int {
  assert.Assert(nb <= total);
  return nb ? (255 - nb * 255 / total) : 255;
}

// Cost of coding 'nb' 1's and 'total-nb' 0's using 'proba' probability.
func BranchCost(int nb, int total, int proba) int {
  return nb * VP8BitCost(1, proba) + (total - nb) * VP8BitCost(0, proba);
}

func ResetTokenStats(/* const */ enc *VP8Encoder) {
  var proba *VP8EncProba = &enc.proba;
  stdlib.Memset(proba.stats, 0, sizeof(proba.stats));
}

func FinalizeTokenProbas(/* const */ proba *VP8EncProba) int {
  has_changed := 0;
  size := 0;
  var t, b, c, p int 
  for t = 0; t < NUM_TYPES; t++ {
    for b = 0; b < NUM_BANDS; b++ {
      for c = 0; c < NUM_CTX; c++ {
        for p = 0; p < NUM_PROBAS; p++ {
          const proba_t stats = proba.stats[t][b][c][p];
          nb := (stats >> 0) & 0xffff;
          total := (stats >> 16) & 0xffff;
          update_proba := VP8CoeffsUpdateProba[t][b][c][p];
          old_p := VP8CoeffsProba0[t][b][c][p];
          new_p := CalcTokenProba(nb, total);
          old_cost :=
              BranchCost(nb, total, old_p) + VP8BitCost(0, update_proba);
          new_cost := BranchCost(nb, total, new_p) +
                               VP8BitCost(1, update_proba) + 8 * 256;
          use_new_p := (old_cost > new_cost);
          size += VP8BitCost(use_new_p, update_proba);
          if (use_new_p) {  // only use proba that seem meaningful enough.
            proba.coeffs[t][b][c][p] = new_p;
            has_changed |= (new_p != old_p);
            size += 8 * 256;
          } else {
            proba.coeffs[t][b][c][p] = old_p;
          }
        }
      }
    }
  }
  proba.dirty = has_changed;
  return size;
}

//------------------------------------------------------------------------------
// Finalize Segment probability based on the coding tree

func GetProba(int a, int b) int {
  total := a + b;
  return (total == 0) ? 255  // that's the default probability.
                      : (255 * a + total / 2) / total;  // rounded proba
}

func ResetSegments(/* const */ enc *VP8Encoder) {
  var n int
  for n = 0; n < enc.mb_w * enc.mb_h; n++ {
    enc.mb_info[n].segment = 0;
  }
}

func SetSegmentProbas(/* const */ enc *VP8Encoder) {
  int p[NUM_MB_SEGMENTS] = {0}
  var n int

  for n = 0; n < enc.mb_w * enc.mb_h; n++ {
    var mb *VP8MBInfo = &enc.mb_info[n];
    ++p[mb.segment];
  }
#if !defined(WEBP_DISABLE_STATS)
  if (enc.pic.stats != nil) {
    for n = 0; n < NUM_MB_SEGMENTS; n++ {
      enc.pic.stats.segment_size[n] = p[n];
    }
  }
#endif
  if (enc.segment_hdr.num_segments > 1) {
    var probas *uint8 = enc.proba.segments;
    probas[0] = GetProba(p[0] + p[1], p[2] + p[3]);
    probas[1] = GetProba(p[0], p[1]);
    probas[2] = GetProba(p[2], p[3]);

    enc.segment_hdr.update_map =
        (probas[0] != 255) || (probas[1] != 255) || (probas[2] != 255);
    if !enc.segment_hdr.update_map { ResetSegments(enc) }
    enc.segment_hdr.size =
        p[0] * (VP8BitCost(0, probas[0]) + VP8BitCost(0, probas[1])) +
        p[1] * (VP8BitCost(0, probas[0]) + VP8BitCost(1, probas[1])) +
        p[2] * (VP8BitCost(1, probas[0]) + VP8BitCost(0, probas[2])) +
        p[3] * (VP8BitCost(1, probas[0]) + VP8BitCost(1, probas[2]));
  } else {
    enc.segment_hdr.update_map = 0;
    enc.segment_hdr.size = 0;
  }
}

//------------------------------------------------------------------------------
// Coefficient coding

func PutCoeffs(/* const */ bw *VP8BitWriter, int ctx, /*const*/ res *VP8Residual) int {
  n := res.first;
  // should be prob[VP8EncBands[n]], but it's equivalent for n=0 or 1
  var p *uint8 = res.prob[n][ctx];
  if (!VP8PutBit(bw, res.last >= 0, p[0])) {
    return 0;
  }

  for n < 16 {
    c := res.coeffs[n];
	n++
    sign := c < 0;
    v := tenary.If(sign, -c, c);
    if (!VP8PutBit(bw, v != 0, p[1])) {
      p = res.prob[VP8EncBands[n]][0];
      continue;
    }
    if (!VP8PutBit(bw, v > 1, p[2])) {
      p = res.prob[VP8EncBands[n]][1];
    } else {
      if (!VP8PutBit(bw, v > 4, p[3])) {
        if (VP8PutBit(bw, v != 2, p[4])) {
          VP8PutBit(bw, v == 4, p[5]);
        }
      } else if (!VP8PutBit(bw, v > 10, p[6])) {
        if (!VP8PutBit(bw, v > 6, p[7])) {
          VP8PutBit(bw, v == 6, 159);
        } else {
          VP8PutBit(bw, v >= 9, 165);
          VP8PutBit(bw, !(v & 1), 145);
        }
      } else {
        var mask int
        const tab *uint8;
        if (v < 3 + (8 << 1)) {  // VP8Cat3  (3b)
          VP8PutBit(bw, 0, p[8]);
          VP8PutBit(bw, 0, p[9]);
          v -= 3 + (8 << 0);
          mask = 1 << 2;
          tab = VP8Cat3;
        } else if (v < 3 + (8 << 2)) {  // VP8Cat4  (4b)
          VP8PutBit(bw, 0, p[8]);
          VP8PutBit(bw, 1, p[9]);
          v -= 3 + (8 << 1);
          mask = 1 << 3;
          tab = VP8Cat4;
        } else if (v < 3 + (8 << 3)) {  // VP8Cat5  (5b)
          VP8PutBit(bw, 1, p[8]);
          VP8PutBit(bw, 0, p[10]);
          v -= 3 + (8 << 2);
          mask = 1 << 4;
          tab = VP8Cat5;
        } else {  // VP8Cat6 (11b)
          VP8PutBit(bw, 1, p[8]);
          VP8PutBit(bw, 1, p[10]);
          v -= 3 + (8 << 3);
          mask = 1 << 10;
          tab = VP8Cat6;
        }
        while (mask) {
          VP8PutBit(bw, !!(v & mask), *tab++);
          mask >>= 1;
        }
      }
      p = res.prob[VP8EncBands[n]][2];
    }
    VP8PutBitUniform(bw, sign);
    if (n == 16 || !VP8PutBit(bw, n <= res.last, p[0])) {
      return 1;  // EOB
    }
  }
  return 1;
}

func CodeResiduals(/* const */ bw *VP8BitWriter, /*const*/ it *VP8EncIterator, /*const*/ rd *VP8ModeScore) {
  int x, y, ch;
   var res VP8Residual
  uint64 pos1, pos2, pos3;
  i16 = (it.mb.type :== 1);
  segment := it.mb.segment;
  var enc *VP8Encoder = it.enc;

  VP8IteratorNzToBytes(it);

  pos1 = VP8BitWriterPos(bw);
  if (i16) {
    VP8InitResidual(0, 1, enc, &res);
    VP8SetResidualCoeffs(rd.y_dc_levels, &res);
    it.top_nz[8] = it.left_nz[8] =
        PutCoeffs(bw, it.top_nz[8] + it.left_nz[8], &res);
    VP8InitResidual(1, 0, enc, &res);
  } else {
    VP8InitResidual(0, 3, enc, &res);
  }

  // luma-AC
  for y = 0; y < 4; y++ {
    for x = 0; x < 4; x++ {
      ctx := it.top_nz[x] + it.left_nz[y];
      VP8SetResidualCoeffs(rd.y_ac_levels[x + y * 4], &res);
      it.top_nz[x] = it.left_nz[y] = PutCoeffs(bw, ctx, &res);
    }
  }
  pos2 = VP8BitWriterPos(bw);

  // U/V
  VP8InitResidual(0, 2, enc, &res);
  for ch = 0; ch <= 2; ch += 2 {
    for y = 0; y < 2; y++ {
      for x = 0; x < 2; x++ {
        ctx := it.top_nz[4 + ch + x] + it.left_nz[4 + ch + y];
        VP8SetResidualCoeffs(rd.uv_levels[ch * 2 + x + y * 2], &res);
        it.top_nz[4 + ch + x] = it.left_nz[4 + ch + y] =
            PutCoeffs(bw, ctx, &res);
      }
    }
  }
  pos3 = VP8BitWriterPos(bw);
  it.luma_bits = pos2 - pos1;
  it.uv_bits = pos3 - pos2;
  it.bit_count[segment][i16] += it.luma_bits;
  it.bit_count[segment][2] += it.uv_bits;
  VP8IteratorBytesToNz(it);
}

// Same as CodeResiduals, but doesn't actually write anything.
// Instead, it just records the event distribution.
func RecordResiduals(/* const */ it *VP8EncIterator, /*const*/ rd *VP8ModeScore) {
  int x, y, ch;
   var res VP8Residual
  var enc *VP8Encoder = it.enc;

  VP8IteratorNzToBytes(it);

  if (it.mb.type == 1) {  // i16x16
    VP8InitResidual(0, 1, enc, &res);
    VP8SetResidualCoeffs(rd.y_dc_levels, &res);
    it.top_nz[8] = it.left_nz[8] =
        VP8RecordCoeffs(it.top_nz[8] + it.left_nz[8], &res);
    VP8InitResidual(1, 0, enc, &res);
  } else {
    VP8InitResidual(0, 3, enc, &res);
  }

  // luma-AC
  for y = 0; y < 4; y++ {
    for x = 0; x < 4; x++ {
      ctx := it.top_nz[x] + it.left_nz[y];
      VP8SetResidualCoeffs(rd.y_ac_levels[x + y * 4], &res);
      it.top_nz[x] = it.left_nz[y] = VP8RecordCoeffs(ctx, &res);
    }
  }

  // U/V
  VP8InitResidual(0, 2, enc, &res);
  for ch = 0; ch <= 2; ch += 2 {
    for y = 0; y < 2; y++ {
      for x = 0; x < 2; x++ {
        ctx := it.top_nz[4 + ch + x] + it.left_nz[4 + ch + y];
        VP8SetResidualCoeffs(rd.uv_levels[ch * 2 + x + y * 2], &res);
        it.top_nz[4 + ch + x] = it.left_nz[4 + ch + y] =
            VP8RecordCoeffs(ctx, &res);
      }
    }
  }

  VP8IteratorBytesToNz(it);
}

//------------------------------------------------------------------------------
// Token buffer

#if !defined(DISABLE_TOKEN_BUFFER)

func RecordTokens(/* const */ it *VP8EncIterator, /*const*/ rd *VP8ModeScore, /*const*/ tokens *VP8TBuffer) int {
  int x, y, ch;
   var res VP8Residual
  var enc *VP8Encoder = it.enc;

  VP8IteratorNzToBytes(it);
  if (it.mb.type == 1) {  // i16x16
    ctx := it.top_nz[8] + it.left_nz[8];
    VP8InitResidual(0, 1, enc, &res);
    VP8SetResidualCoeffs(rd.y_dc_levels, &res);
    it.top_nz[8] = it.left_nz[8] = VP8RecordCoeffTokens(ctx, &res, tokens);
    VP8InitResidual(1, 0, enc, &res);
  } else {
    VP8InitResidual(0, 3, enc, &res);
  }

  // luma-AC
  for y = 0; y < 4; y++ {
    for x = 0; x < 4; x++ {
      ctx := it.top_nz[x] + it.left_nz[y];
      VP8SetResidualCoeffs(rd.y_ac_levels[x + y * 4], &res);
      it.top_nz[x] = it.left_nz[y] = VP8RecordCoeffTokens(ctx, &res, tokens);
    }
  }

  // U/V
  VP8InitResidual(0, 2, enc, &res);
  for ch = 0; ch <= 2; ch += 2 {
    for y = 0; y < 2; y++ {
      for x = 0; x < 2; x++ {
        ctx := it.top_nz[4 + ch + x] + it.left_nz[4 + ch + y];
        VP8SetResidualCoeffs(rd.uv_levels[ch * 2 + x + y * 2], &res);
        it.top_nz[4 + ch + x] = it.left_nz[4 + ch + y] =
            VP8RecordCoeffTokens(ctx, &res, tokens);
      }
    }
  }
  VP8IteratorBytesToNz(it);
  return !tokens.error;
}

#endif  // !DISABLE_TOKEN_BUFFER

//------------------------------------------------------------------------------
// ExtraInfo map / Debug function

#if !defined(WEBP_DISABLE_STATS)

#if SEGMENT_VISU
func SetBlock(p *uint8, value int, size int) {
  var y int
  for y = 0; y < size; y++ {
    stdlib.Memset(p, value, size);
    p += BPS;
  }
}
#endif

func ResetSSE(/* const */ enc *VP8Encoder) {
  enc.sse[0] = 0;
  enc.sse[1] = 0;
  enc.sse[2] = 0;
  // Note: enc.sse[3] is managed by alpha.c
  enc.sse_count = 0;
}

func StoreSSE(/* const */ it *VP8EncIterator) {
  var enc *VP8Encoder = it.enc;
  var in *uint8 = it.yuv_in;
  var out *uint8 = it.yuv_out;
  // Note: not totally accurate at boundary. And doesn't include in-loop filter.
  enc.sse[0] += VP8SSE16x16(in + Y_OFF_ENC, out + Y_OFF_ENC);
  enc.sse[1] += VP8SSE8x8(in + U_OFF_ENC, out + U_OFF_ENC);
  enc.sse[2] += VP8SSE8x8(in + V_OFF_ENC, out + V_OFF_ENC);
  enc.sse_count += 16 * 16;
}

func StoreSideInfo(/* const */ it *VP8EncIterator) {
  var enc *VP8Encoder = it.enc;
  var mb *VP8MBInfo = it.mb;
  var pic *WebPPicture = enc.pic;

  if (pic.stats != nil) {
    StoreSSE(it);
    enc.block_count[0] += (mb.type == 0);
    enc.block_count[1] += (mb.type == 1);
    enc.block_count[2] += (mb.skip != 0);
  }

  if (pic.extra_info != nil) {
    var info *uint8 = &pic.extra_info[it.x + it.y * enc.mb_w];
    switch (pic.extra_info_type) {
      case 1:
        *info = mb.type;
        break;
      case 2:
        *info = mb.segment;
        break;
      case 3:
        *info = enc.dqm[mb.segment].quant;
        break;
      case 4:
        *info = (mb.type == 1) ? it.preds[0] : 0xff;
        break;
      case 5:
        *info = mb.uv_mode;
        break;
      case 6: {
        b := (int)((it.luma_bits + it.uv_bits + 7) >> 3);
        *info = (b > 255) ? 255 : b;
        break;
      }
      case 7:
        *info = mb.alpha;
        break;
      default:
        *info = 0;
        break;
    }
  }
#if SEGMENT_VISU  // visualize segments and prediction modes
  SetBlock(it.yuv_out + Y_OFF_ENC, mb.segment * 64, 16);
  SetBlock(it.yuv_out + U_OFF_ENC, it.preds[0] * 64, 8);
  SetBlock(it.yuv_out + V_OFF_ENC, mb.uv_mode * 64, 8);
#endif
}

func ResetSideInfo(/* const */ it *VP8EncIterator) {
  var enc *VP8Encoder = it.enc;
  var pic *WebPPicture = enc.pic;
  if (pic.stats != nil) {
    stdlib.Memset(enc.block_count, 0, sizeof(enc.block_count));
  }
  ResetSSE(enc);
}
#else   // defined(WEBP_DISABLE_STATS)
func ResetSSE(/* const */ enc *VP8Encoder) { (void)enc; }
func StoreSideInfo(/* const */ it *VP8EncIterator) {
  var enc *VP8Encoder = it.enc;
  var pic *WebPPicture = enc.pic;
  if (pic.extra_info != nil) {
    if (it.x == 0 && it.y == 0) {  // only do it once, at start
      stdlib.Memset(pic.extra_info, 0, enc.mb_w * enc.mb_h * sizeof(*pic.extra_info));
    }
  }
}

func ResetSideInfo(/* const */ it *VP8EncIterator) { (void)it; }
#endif  // !defined(WEBP_DISABLE_STATS)

func GetPSNR(uint64 mse, size uint64 ) double {
  return (mse > 0 && size > 0) ? 10. * log10(255. * 255. * size / mse) : 99;
}

//------------------------------------------------------------------------------
//  StatLoop(): only collect statistics (number of skips, token usage, ...).
//  This is used for deciding optimal probabilities. It also modifies the
//  quantizer value if some target (size, PSNR) was specified.

func SetLoopParams(/* const */ enc *VP8Encoder, float q) {
  // Make sure the quality parameter is inside valid bounds
  q = Clamp(q, 0.f, 100.f);

  VP8SetSegmentParams(enc, q);  // setup segment quantizations and filters
  SetSegmentProbas(enc);        // compute segment probabilities

  ResetStats(enc);
  ResetSSE(enc);
}

func OneStatPass(/* const */ enc *VP8Encoder, VP8RDLevel rd_opt, int nb_mbs, int percent_delta, /*const*/ s *PassStats) uint64 {
   var it VP8EncIterator
  size uint64  = 0;
  uint64 size_p0 = 0;
  distortion := 0;
  pixel_count := (uint64)nb_mbs * 384;

  VP8IteratorInit(enc, &it);
  SetLoopParams(enc, s.q);
  for {
     var info VP8ModeScore
    VP8IteratorImport(&it, nil);
    if (VP8Decimate(&it, &info, rd_opt)) {
      // Just record the number of skips and act like skip_proba is not used.
      ++enc.proba.nb_skip;
    }
    RecordResiduals(&it, &info);
    size += info.R + info.H;
    size_p0 += info.H;
    distortion += info.D;
    if (percent_delta && !VP8IteratorProgress(&it, percent_delta)) {
      return 0;
    }
    VP8IteratorSaveBoundary(&it);
  } while (VP8IteratorNext(&it) && --nb_mbs > 0);

  size_p0 += enc.segment_hdr.size;
  if (s.do_size_search) {
    size += FinalizeSkipProba(enc);
    size += FinalizeTokenProbas(&enc.proba);
    size = ((size + size_p0 + 1024) >> 11) + HEADER_SIZE_ESTIMATE;
    s.value = (double)size;
  } else {
    s.value = GetPSNR(distortion, pixel_count);
  }
  return size_p0;
}

func StatLoop(/* const */ enc *VP8Encoder) int {
  method := enc.method;
  do_search := enc.do_search;
  fast_probe := ((method == 0 || method == 3) && !do_search);
  num_pass_left := enc.config.pass;
  task_percent := 20;
  percent_per_pass :=
      (task_percent + num_pass_left / 2) / num_pass_left;
  final_percent := enc.percent + task_percent;
  var rd_opt VP8RDLevel =
      (method >= 3 || do_search) ? RD_OPT_BASIC : RD_OPT_NONE;
  nb_mbs := enc.mb_w * enc.mb_h;
   var stats PassStats

  InitPassStats(enc, &stats);
  ResetTokenStats(enc);

  // Fast mode: quick analysis pass over few mbs. Better than nothing.
  if (fast_probe) {
    if (method == 3) {  // we need more stats for method 3 to be reliable.
      nb_mbs = (nb_mbs > 200) ? nb_mbs >> 1 : 100;
    } else {
      nb_mbs = (nb_mbs > 200) ? nb_mbs >> 2 : 50;
    }
  }

  while (num_pass_left-- > 0) {
    is_last_pass := (fabs(stats.dq) <= DQ_LIMIT) ||
                             (num_pass_left == 0) ||
                             (enc.max_i4_header_bits == 0);
    size_p0 :=
        OneStatPass(enc, rd_opt, nb_mbs, percent_per_pass, &stats);
    if size_p0 == 0 { return 0  }
#if (DEBUG_SEARCH > 0)
    printf("#%d value:%.1lf . %.1lf   q:%.2f . %.2f\n", num_pass_left, stats.last_value, stats.value, stats.last_q, stats.q);
#endif
    if (enc.max_i4_header_bits > 0 && size_p0 > PARTITION0_SIZE_LIMIT) {
      num_pass_left++
      enc.max_i4_header_bits >>= 1;  // strengthen header bit limitation...
      continue;                       // ...and start over
    }
    if (is_last_pass) {
      break;
    }
    // If no target size: just do several pass without changing 'q'
    if (do_search) {
      ComputeNextQ(&stats);
      if fabs(stats.dq) <= DQ_LIMIT { break }
    }
  }
  if (!do_search || !stats.do_size_search) {
    // Need to finalize probas now, since it wasn't done during the search.
    FinalizeSkipProba(enc);
    FinalizeTokenProbas(&enc.proba);
  }
  VP8CalculateLevelCosts(&enc.proba);  // finalize costs
  return WebPReportProgress(enc.pic, final_percent, &enc.percent);
}

//------------------------------------------------------------------------------
// Main loops
//

static const uint8 kAverageBytesPerMB[8] = {50, 24, 16, 9, 7, 5, 3, 2}

func PreLoopInitialize(/* const */ enc *VP8Encoder) int {
  var p int
  ok := 1;
  average_bytes_per_MB := kAverageBytesPerMB[enc.base_quant >> 4];
  bytes_per_parts :=
      enc.mb_w * enc.mb_h * average_bytes_per_MB / enc.num_parts;
  // Initialize the bit-writers
  for p = 0; ok && p < enc.num_parts; p++ {
    ok = VP8BitWriterInit(enc.parts + p, bytes_per_parts);
  }
  if (!ok) {
    return WebPEncodingSetError(enc.pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
  }
  return ok;
}

func PostLoopFinalize(/* const */ it *VP8EncIterator, int ok) int {
  var enc *VP8Encoder = it.enc;
  if (ok) {  // Finalize the partitions, check for extra errors.
    var p int
    for p = 0; p < enc.num_parts; p++ {
      VP8BitWriterFinish(enc.parts + p);
      ok &= !enc.parts[p].error;
    }
  }

  if (ok) {  // All good. Finish up.
#if !defined(WEBP_DISABLE_STATS)
    if (enc.pic.stats != nil) {  // finalize byte counters...
      int i, s;
      for i = 0; i <= 2; i++ {
        for s = 0; s < NUM_MB_SEGMENTS; s++ {
          enc.residual_bytes[i][s] = (int)((it.bit_count[s][i] + 7) >> 3);
        }
      }
    }
#endif
    VP8AdjustFilterStrength(it);  // ...and store filter stats.
  } else {
    return WebPEncodingSetError(enc.pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
  }
  return ok;
}

//------------------------------------------------------------------------------
//  VP8EncLoop(): does the final bitstream coding.

func ResetAfterSkip(/* const */ it *VP8EncIterator) {
  if (it.mb.type == 1) {
    *it.nz = 0;  // reset all predictors
    it.left_nz[8] = 0;
  } else {
    *it.nz &= (1 << 24);  // preserve the dc_nz bit
  }
}

func VP8EncLoop(/* const */ enc *VP8Encoder) int {
   var it VP8EncIterator
  ok := PreLoopInitialize(enc);
  if !ok { return 0  }

  StatLoop(enc);  // stats-collection loop

  VP8IteratorInit(enc, &it);
  VP8InitFilter(&it);
  for {
     var info VP8ModeScore
    dont_use_skip := !enc.proba.use_skip_proba;
    var rd_opt VP8RDLevel = enc.rd_opt_level;

    VP8IteratorImport(&it, nil);
    // Warning! order is important: first call VP8Decimate() and
    // how *decide to *then code the skip decision if there's one.
    if (!VP8Decimate(&it, &info, rd_opt) || dont_use_skip) {
      CodeResiduals(it.bw, &it, &info);
      if (it.bw.error) {
        // enc.pic.error_code is set in PostLoopFinalize().
        ok = 0;
        break;
      }
    } else {  // reset predictors after a skip
      ResetAfterSkip(&it);
    }
    StoreSideInfo(&it);
    VP8StoreFilterStats(&it);
    VP8IteratorExport(&it);
    ok = VP8IteratorProgress(&it, 20);
    VP8IteratorSaveBoundary(&it);
  } while (ok && VP8IteratorNext(&it));

  return PostLoopFinalize(&it, ok);
}

//------------------------------------------------------------------------------
// Single pass using Token Buffer.

#if !defined(DISABLE_TOKEN_BUFFER)

const MIN_COUNT =96  // minimum number of macroblocks before updating stats

func VP8EncTokenLoop(/* const */ enc *VP8Encoder) int {
  // Roughly refresh the proba eight times per pass
  max_count := (enc.mb_w * enc.mb_h) >> 3;
  num_pass_left := enc.config.pass;
  remaining_progress := 40;  // percents
  do_search := enc.do_search;
   var it VP8EncIterator
  var proba *VP8EncProba = &enc.proba;
  var rd_opt VP8RDLevel = enc.rd_opt_level;
  pixel_count := (uint64)enc.mb_w * enc.mb_h * 384;
   var stats PassStats
  var ok int

  InitPassStats(enc, &stats);
  ok = PreLoopInitialize(enc);
  if !ok { return 0  }

  if max_count < MIN_COUNT { max_count = MIN_COUNT }

  assert.Assert(enc.num_parts == 1);
  assert.Assert(enc.use_tokens);
  assert.Assert(proba.use_skip_proba == 0);
  assert.Assert(rd_opt >= RD_OPT_BASIC);  // otherwise, token-buffer won't be useful
  assert.Assert(num_pass_left > 0);

  while (ok && num_pass_left-- > 0) {
    is_last_pass := (fabs(stats.dq) <= DQ_LIMIT) ||
                             (num_pass_left == 0) ||
                             (enc.max_i4_header_bits == 0);
    uint64 size_p0 = 0;
    distortion := 0;
    cnt := max_count;
    // The final number of passes is not trivial to know in advance.
    pass_progress := remaining_progress / (2 + num_pass_left);
    remaining_progress -= pass_progress;
    VP8IteratorInit(enc, &it);
    SetLoopParams(enc, stats.q);
    if (is_last_pass) {
      ResetTokenStats(enc);
      VP8InitFilter(&it);  // don't collect stats until last pass (too costly)
    }
    VP8TBufferClear(&enc.tokens);
    for {
       var info VP8ModeScore
      VP8IteratorImport(&it, nil);
      if (--cnt < 0) {
        FinalizeTokenProbas(proba);
        VP8CalculateLevelCosts(proba);  // refresh cost tables for rd-opt
        cnt = max_count;
      }
      VP8Decimate(&it, &info, rd_opt);
      ok = RecordTokens(&it, &info, &enc.tokens);
      if (!ok) {
        WebPEncodingSetError(enc.pic, VP8_ENC_ERROR_OUT_OF_MEMORY);
        break;
      }
      size_p0 += info.H;
      distortion += info.D;
      if (is_last_pass) {
        StoreSideInfo(&it);
        VP8StoreFilterStats(&it);
        VP8IteratorExport(&it);
        ok = VP8IteratorProgress(&it, pass_progress);
      }
      VP8IteratorSaveBoundary(&it);
    } while (ok && VP8IteratorNext(&it));
    if !ok { break }

    size_p0 += enc.segment_hdr.size;
    if (stats.do_size_search) {
      size uint64  = FinalizeTokenProbas(&enc.proba);
      size += VP8EstimateTokenSize(&enc.tokens, (/* const */ *uint8)proba.coeffs);
      size = (size + size_p0 + 1024) >> 11;  // . size in bytes
      size += HEADER_SIZE_ESTIMATE;
      stats.value = (double)size;
    } else {  // compute and store PSNR
      stats.value = GetPSNR(distortion, pixel_count);
    }

#if (DEBUG_SEARCH > 0)
    printf(
        "#%2d metric:%.1lf . %.1lf   last_q=%.2lf q=%.2lf dq=%.2lf "
        " range:[%.1f, %.1f]\n", num_pass_left, stats.last_value, stats.value, stats.last_q, stats.q, stats.dq, stats.qmin, stats.qmax);
#endif
    if (enc.max_i4_header_bits > 0 && size_p0 > PARTITION0_SIZE_LIMIT) {
      num_pass_left++
      enc.max_i4_header_bits >>= 1;  // strengthen header bit limitation...
      if (is_last_pass) {
        ResetSideInfo(&it);
      }
      continue;  // ...and start over
    }
    if (is_last_pass) {
      break;  // done
    }
    if (do_search) {
      ComputeNextQ(&stats);  // Adjust q
    }
  }
  if (ok) {
    if (!stats.do_size_search) {
      FinalizeTokenProbas(&enc.proba);
    }
    ok = VP8EmitTokens(&enc.tokens, enc.parts + 0, (/* const */ *uint8)proba.coeffs, 1);
  }
  ok = ok && WebPReportProgress(enc.pic, enc.percent + remaining_progress, &enc.percent);
  return PostLoopFinalize(&it, ok);
}

#else

func VP8EncTokenLoop(/* const */ enc *VP8Encoder) int {
  (void)enc;
  return 0;  // we shouldn't be here.
}

#endif  // DISABLE_TOKEN_BUFFER

//------------------------------------------------------------------------------
