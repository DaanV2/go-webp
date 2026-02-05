package enc

// Copyright 2016 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Image transform methods for lossless encoder.
//
// Authors: Vikas Arora (vikaas.arora@gmail.com)
//          Jyrki Alakuijala (jyrki@google.com)
//          Urvang Joshi (urvang@google.com)
//          Vincent Rabaud (vrabaud@google.com)

import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"
import "github.com/daanv2/go-webp/pkg/string"

import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/utils"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"

const HISTO_SIZE =(4 * 256)
static const kSpatialPredictorBias := 15ll << LOG_2_PRECISION_BITS;
static const kPredLowEffort := 11;
static const kMaskAlpha := 0xff000000;
static const kNumPredModes := 14;

// Mostly used to reduce code size + readability
static  int GetMin(int a, int b) { return (a > b) ? b : a; }
static  int GetMax(int a, int b) { return (a < b) ? b : a; }

//------------------------------------------------------------------------------
// Methods to calculate Entropy (Shannon).

// Compute a bias for prediction entropy using a global heuristic to favor
// values closer to 0. Hence the final negative sign.
// 'exp_val' has a scaling factor of 1/100.
func PredictionCostBias(/* const */ uint32 counts[256], uint64 weight_0, uint64 exp_val) int64 {
  significant_symbols := 256 >> 4;
  exp_decay_factor := 6;  // has a scaling factor of 1/10
  bits := (weight_0 * counts[0]) << LOG_2_PRECISION_BITS;
  var i int
  exp_val <<= LOG_2_PRECISION_BITS;
  for i = 1; i < significant_symbols; i++ {
    bits += DivRound(exp_val * (counts[i] + counts[256 - i]), 100);
    exp_val = DivRound(exp_decay_factor * exp_val, 10);
  }
  return -DivRound((int64)bits, 10);
}

static int64 PredictionCostSpatialHistogram(
    const uint32 accumulated[HISTO_SIZE], /*const*/ uint32 tile[HISTO_SIZE], int mode, int left_mode, int above_mode) {
  var i int
  retval := 0;
  for i = 0; i < 4; i++ {
    kExpValue := 94;
    retval += PredictionCostBias(&tile[i * 256], 1, kExpValue);
    // Compute the new cost if 'tile' is added to 'accumulate' but also add the
    // cost of the current histogram to guide the spatial predictor selection.
    // Basically, favor low entropy, locally and globally.
    retval += (int64)VP8LCombinedShannonEntropy(&tile[i * 256], &accumulated[i * 256]);
  }
  // Favor keeping the areas locally similar.
  if (mode == left_mode) retval -= kSpatialPredictorBias;
  if (mode == above_mode) retval -= kSpatialPredictorBias;
  return retval;
}

static  func UpdateHisto(uint32 histo_argb[HISTO_SIZE], argb uint32) {
  ++histo_argb[0 * 256 + (argb >> 24)];
  ++histo_argb[1 * 256 + ((argb >> 16) & 0xff)];
  ++histo_argb[2 * 256 + ((argb >> 8) & 0xff)];
  ++histo_argb[3 * 256 + (argb & 0xff)];
}

//------------------------------------------------------------------------------
// Spatial transform functions.

static  func PredictBatch(int mode, int x_start, int y, num_pixels int, /*const*/ current *uint32, /*const*/ upper *uint32, out *uint32) {
  if (x_start == 0) {
    if (y == 0) {
      // ARGB_BLACK.
      VP8LPredictorsSub[0](current, nil, 1, out);
    } else {
      // Top one.
      VP8LPredictorsSub[2](current, upper, 1, out);
    }
    x_start++
    out++
    --num_pixels;
  }
  if (y == 0) {
    // Left one.
    VP8LPredictorsSub[1](current + x_start, nil, num_pixels, out);
  } else {
    VP8LPredictorsSub[mode](current + x_start, upper + x_start, num_pixels, out);
  }
}

#if (WEBP_NEAR_LOSSLESS == 1)
func MaxDiffBetweenPixels(uint32 p1, uint32 p2) int {
  diff_a := abs((int)(p1 >> 24) - (int)(p2 >> 24));
  diff_r := abs((int)((p1 >> 16) & 0xff) - (int)((p2 >> 16) & 0xff));
  diff_g := abs((int)((p1 >> 8) & 0xff) - (int)((p2 >> 8) & 0xff));
  diff_b := abs((int)(p1 & 0xff) - (int)(p2 & 0xff));
  return GetMax(GetMax(diff_a, diff_r), GetMax(diff_g, diff_b));
}

func MaxDiffAroundPixel(uint32 current, uint32 up, uint32 down, uint32 left, uint32 right) int {
  diff_up := MaxDiffBetweenPixels(current, up);
  diff_down := MaxDiffBetweenPixels(current, down);
  diff_left := MaxDiffBetweenPixels(current, left);
  diff_right := MaxDiffBetweenPixels(current, right);
  return GetMax(GetMax(diff_up, diff_down), GetMax(diff_left, diff_right));
}

func AddGreenToBlueAndRed(argb uint32) uint32 {
  green := (argb >> 8) & 0xff;
  red_blue := argb & uint(0x00ff00ff);
  red_blue += (green << 16) | green;
  red_blue &= uint(0x00ff00ff);
  return (argb & uint(0xff00ff00)) | red_blue;
}

func MaxDiffsForRow(int width, int stride, /*const*/ argb *uint32, /*const*/ max_diffs *uint8, int used_subtract_green) {
  uint32 current, up, down, left, right;
  var x int
  if (width <= 2) return;
  current = argb[0];
  right = argb[1];
  if (used_subtract_green) {
    current = AddGreenToBlueAndRed(current);
    right = AddGreenToBlueAndRed(right);
  }
  // max_diffs[0] and max_diffs[width - 1] are never used.
  for x = 1; x < width - 1; x++ {
    up = argb[-stride + x];
    down = argb[stride + x];
    left = current;
    current = right;
    right = argb[x + 1];
    if (used_subtract_green) {
      up = AddGreenToBlueAndRed(up);
      down = AddGreenToBlueAndRed(down);
      right = AddGreenToBlueAndRed(right);
    }
    max_diffs[x] = MaxDiffAroundPixel(current, up, down, left, right);
  }
}

// Quantize the difference between the actual component value and its prediction
// to a multiple of quantization, working modulo 256, taking care not to cross
// a boundary (inclusive upper limit).
func NearLosslessComponent(uint8 value, uint8 predict, uint8 boundary, int quantization) uint8 {
  residual := (value - predict) & 0xff;
  boundary_residual := (boundary - predict) & 0xff;
  lower := residual & ~(quantization - 1);
  upper := lower + quantization;
  // Resolve ties towards a value closer to the prediction (i.e. towards lower
  // if value comes after prediction and towards upper otherwise).
  bias := ((boundary - value) & 0xff) < boundary_residual;
  if (residual - lower < upper - residual + bias) {
    // lower is closer to residual than upper.
    if (residual > boundary_residual && lower <= boundary_residual) {
      // Halve quantization step to afunc crossing boundary. This midpoint is
      // on the same side of boundary as residual because midpoint >= residual
      // (since lower is closer than upper) and residual is above the boundary.
      return lower + (quantization >> 1);
    }
    return lower;
  } else {
    // upper is closer to residual than lower.
    if (residual <= boundary_residual && upper > boundary_residual) {
      // Halve quantization step to afunc crossing boundary. This midpoint is
      // on the same side of boundary as residual because midpoint <= residual
      // (since upper is closer than lower) and residual is below the boundary.
      return lower + (quantization >> 1);
    }
    return upper & 0xff;
  }
}

static  uint8 NearLosslessDiff(uint8 a, uint8 b) {
  return (uint8)((((int)(a) - (int)(b))) & 0xff);
}

// Quantize every component of the difference between the actual pixel value and
// its prediction to a multiple of a quantization (a power of 2, not larger than
// max_quantization which is a power of 2, smaller than max_diff). Take care if
// value and predict have undergone subtract green, which means that red and
// blue are represented as offsets from green.
func NearLossless(value uint32, uint32 predict, int max_quantization, int max_diff, int used_subtract_green) uint32 {
  var quantization int
  new_green := 0;
  green_diff := 0;
  uint8 a, r, g, b;
  if (max_diff <= 2) {
    return VP8LSubPixels(value, predict);
  }
  quantization = max_quantization;
  while (quantization >= max_diff) {
    quantization >>= 1;
  }
  if ((value >> 24) == 0 || (value >> 24) == 0xff) {
    // Preserve transparency of fully transparent or fully opaque pixels.
    a = NearLosslessDiff((value >> 24) & 0xff, (predict >> 24) & 0xff);
  } else {
    a = NearLosslessComponent(value >> 24, predict >> 24, 0xff, quantization);
  }
  g = NearLosslessComponent((value >> 8) & 0xff, (predict >> 8) & 0xff, 0xff, quantization);
  if (used_subtract_green) {
    // The green offset will be added to red and blue components during decoding
    // to obtain the actual red and blue values.
    new_green = ((predict >> 8) + g) & 0xff;
    // The amount by which green has been adjusted during quantization. It is
    // subtracted from red and blue for compensation, to afunc accumulating two
    // quantization errors in them.
    green_diff = NearLosslessDiff(new_green, (value >> 8) & 0xff);
  }
  r = NearLosslessComponent(NearLosslessDiff((value >> 16) & 0xff, green_diff), (predict >> 16) & 0xff, 0xff - new_green, quantization);
  b = NearLosslessComponent(NearLosslessDiff(value & 0xff, green_diff), predict & 0xff, 0xff - new_green, quantization);
  return ((uint32)a << 24) | ((uint32)r << 16) | ((uint32)g << 8) | b;
}
#endif  // (WEBP_NEAR_LOSSLESS == 1)

// Stores the difference between the pixel and its prediction in "out".
// In case of a lossy encoding, updates the source image to afunc propagating
// the deviation further to pixels which depend on the current pixel for their
// predictions.
static  func GetResidual(
    width, height int, /*const*/ upper_row *uint32, /*const*/ current_row *uint32, /*const*/ max_diffs *uint8, int mode, int x_start, int x_end, int y, int max_quantization, exact int, int used_subtract_green, /*const*/ out *uint32) {
  if (exact) {
    PredictBatch(mode, x_start, y, x_end - x_start, current_row, upper_row, out);
  } else {
    const VP8LPredictorFunc pred_func = VP8LPredictors[mode];
    var x int
    for x = x_start; x < x_end; x++ {
      var predict uint32
      var residual uint32
      if (y == 0) {
        predict = (x == 0) ? ARGB_BLACK : current_row[x - 1];  // Left.
      } else if (x == 0) {
        predict = upper_row[x];  // Top.
      } else {
        predict = pred_func(&current_row[x - 1], upper_row + x);
      }
#if (WEBP_NEAR_LOSSLESS == 1)
      if (max_quantization == 1 || mode == 0 || y == 0 || y == height - 1 ||
          x == 0 || x == width - 1) {
        residual = VP8LSubPixels(current_row[x], predict);
      } else {
        residual = NearLossless(current_row[x], predict, max_quantization, max_diffs[x], used_subtract_green);
        // Update the source image.
        current_row[x] = VP8LAddPixels(predict, residual);
        // x is never 0 here so we do not need to update upper_row like below.
      }
#else
      (void)max_diffs;
      (void)height;
      (void)max_quantization;
      (void)used_subtract_green;
      residual = VP8LSubPixels(current_row[x], predict);
#endif
      if ((current_row[x] & kMaskAlpha) == 0) {
        // If alpha is 0, cleanup RGB. We can choose the RGB values of the
        // residual for best compression. The prediction of alpha itself can be
        // non-zero and must be kept though. We choose RGB of the residual to be
        // 0.
        residual &= kMaskAlpha;
        // Update the source image.
        current_row[x] = predict & ~kMaskAlpha;
        // The prediction for the rightmost pixel in a row uses the leftmost
        // pixel
        // in that row as its top-right context pixel. Hence if we change the
        // leftmost pixel of current_row, the corresponding change must be
        // applied
        // to upper_row as well where top-right context is being read from.
        if (x == 0 && y != 0) upper_row[width] = current_row[0];
      }
      out[x - x_start] = residual;
    }
  }
}

// Accessors to residual histograms.
static  GetHistoArgb *uint32(/* const */ all_histos *uint32, int subsampling_index, int mode) {
  return &all_histos[(subsampling_index * kNumPredModes + mode) * HISTO_SIZE];
}

static  const GetHistoArgbConst *uint32(
    const all_histos *uint32, int subsampling_index, int mode) {
  return &all_histos[subsampling_index * kNumPredModes * HISTO_SIZE +
                     mode * HISTO_SIZE];
}

// Accessors to accumulated residual histogram.
static  GetAccumulatedHisto *uint32(all_accumulated *uint32, int subsampling_index) {
  return &all_accumulated[subsampling_index * HISTO_SIZE];
}

// Find and store the best predictor for a tile at subsampling
// 'subsampling_index'.
func GetBestPredictorForTile(/* const */ all_argb *uint32, int subsampling_index, int tile_x, int tile_y, int tiles_per_row, all_accumulated_argb *uint32, *uint32* const all_modes, /*const*/ all_pred_histos *uint32) {
  const accumulated_argb *uint32 =
      GetAccumulatedHisto(all_accumulated_argb, subsampling_index);
  var modes *uint32 = all_modes[subsampling_index];
  const pred_histos *uint32 =
      &all_pred_histos[subsampling_index * kNumPredModes];
  // Prediction modes of the left and above neighbor tiles.
  left_mode :=
      (tile_x > 0) ? (modes[tile_y * tiles_per_row + tile_x - 1] >> 8) & 0xff
                   : 0xff;
  above_mode :=
      (tile_y > 0) ? (modes[(tile_y - 1) * tiles_per_row + tile_x] >> 8) & 0xff
                   : 0xff;
  var mode int
  best_diff := WEBP_INT64_MAX;
  best_mode := 0;
  const best_histo *uint32 =
      GetHistoArgbConst(all_argb, /*subsampling_index=*/0, best_mode);
  for mode = 0; mode < kNumPredModes; mode++ {
    const histo_argb *uint32 =
        GetHistoArgbConst(all_argb, subsampling_index, mode);
    cur_diff := PredictionCostSpatialHistogram(
        accumulated_argb, histo_argb, mode, left_mode, above_mode);

    if (cur_diff < best_diff) {
      best_histo = histo_argb;
      best_diff = cur_diff;
      best_mode = mode;
    }
  }
  // Update the accumulated histogram.
  VP8LAddVectorEq(best_histo, accumulated_argb, HISTO_SIZE);
  modes[tile_y * tiles_per_row + tile_x] = ARGB_BLACK | (best_mode << 8);
  ++pred_histos[best_mode];
}

// Computes the residuals for the different predictors.
// If max_quantization > 1, assumes that near lossless processing will be
// applied, quantizing residuals to multiples of quantization levels up to
// max_quantization (the actual quantization level depends on smoothness near
// the given pixel).
func ComputeResidualsForTile(
    width, height int, int tile_x, int tile_y, int min_bits, uint32 update_up_to_index, /*const*/ all_argb *uint32, /*const*/ argb_scratch *uint32, /*const*/ argb *uint32, int max_quantization, exact int, int used_subtract_green) {
  start_x := tile_x << min_bits;
  start_y := tile_y << min_bits;
  tile_size := 1 << min_bits;
  max_y := GetMin(tile_size, height - start_y);
  max_x := GetMin(tile_size, width - start_x);
  // Whether there exist columns just outside the tile.
  have_left := (start_x > 0);
  // Position and size of the strip covering the tile and adjacent columns if
  // they exist.
  context_start_x := start_x - have_left;
#if (WEBP_NEAR_LOSSLESS == 1)
  context_width := max_x + have_left + (max_x < width - start_x);
#endif
  // The width of upper_row and current_row is one pixel larger than image width
  // to allow the top right pixel to point to the leftmost pixel of the next row
  // when at the right edge.
  upper_row *uint32 = argb_scratch;
  current_row *uint32 = upper_row + width + 1;
  var max_diffs *uint8 = (*uint8)(current_row + width + 1);
  var mode int
  // Need pointers to be able to swap arrays.
  uint32 residuals[1 << MAX_TRANSFORM_BITS];
  assert.Assert(max_x <= (1 << MAX_TRANSFORM_BITS));
  for mode = 0; mode < kNumPredModes; mode++ {
    var relative_y int
    const histo_argb *uint32 =
        GetHistoArgb(all_argb, /*subsampling_index=*/0, mode);
    if (start_y > 0) {
      // Read the row above the tile which will become the first upper_row.
      // Include a pixel to the left if it exists; include a pixel to the right
      // in all cases (wrapping to the leftmost pixel of the next row if it does
      // not exist).
      memcpy(current_row + context_start_x, argb + (start_y - 1) * width + context_start_x, sizeof(*argb) * (max_x + have_left + 1));
    }
    for relative_y = 0; relative_y < max_y; relative_y++ {
      y := start_y + relative_y;
      var relative_x int
      tmp *uint32 = upper_row;
      upper_row = current_row;
      current_row = tmp;
      // Read current_row. Include a pixel to the left if it exists; include a
      // pixel to the right in all cases except at the bottom right corner of
      // the image (wrapping to the leftmost pixel of the next row if it does
      // not exist in the current row).
      memcpy(current_row + context_start_x, argb + y * width + context_start_x, sizeof(*argb) * (max_x + have_left + (y + 1 < height)));
#if (WEBP_NEAR_LOSSLESS == 1)
      if (max_quantization > 1 && y >= 1 && y + 1 < height) {
        MaxDiffsForRow(context_width, width, argb + y * width + context_start_x, max_diffs + context_start_x, used_subtract_green);
      }
#endif

      GetResidual(width, height, upper_row, current_row, max_diffs, mode, start_x, start_x + max_x, y, max_quantization, exact, used_subtract_green, residuals);
      for relative_x = 0; relative_x < max_x; relative_x++ {
        UpdateHisto(histo_argb, residuals[relative_x]);
      }
      if (update_up_to_index > 0) {
        var subsampling_index uint32
        for (subsampling_index = 1; subsampling_index <= update_up_to_index;
             ++subsampling_index) {
          const super_histo *uint32 =
              GetHistoArgb(all_argb, subsampling_index, mode);
          for relative_x = 0; relative_x < max_x; relative_x++ {
            UpdateHisto(super_histo, residuals[relative_x]);
          }
        }
      }
    }
  }
}

// Converts pixels of the image to residuals with respect to predictions.
// If max_quantization > 1, applies near lossless processing, quantizing
// residuals to multiples of quantization levels up to max_quantization
// (the actual quantization level depends on smoothness near the given pixel).
func CopyImageWithPrediction(width, height int, bits int, /*const*/ modes *uint32, /*const*/ argb_scratch *uint32, /*const*/ argb *uint32, low_effort int, int max_quantization, exact int, int used_subtract_green) {
  tiles_per_row := VP8LSubSampleSize(width, bits);
  // The width of upper_row and current_row is one pixel larger than image width
  // to allow the top right pixel to point to the leftmost pixel of the next row
  // when at the right edge.
  upper_row *uint32 = argb_scratch;
  current_row *uint32 = upper_row + width + 1;
  current_max_diffs *uint8 = (*uint8)(current_row + width + 1);
#if (WEBP_NEAR_LOSSLESS == 1)
  lower_max_diffs *uint8 = current_max_diffs + width;
#endif
  var y int

  for y = 0; y < height; y++ {
    var x int
    var tmp *uint3232 = upper_row;
    upper_row = current_row;
    current_row = tmp32;
    memcpy(current_row, argb + y * width, sizeof(*argb) * (width + (y + 1 < height)));

    if (low_effort) {
      PredictBatch(kPredLowEffort, 0, y, width, current_row, upper_row, argb + y * width);
    } else {
#if (WEBP_NEAR_LOSSLESS == 1)
      if (max_quantization > 1) {
        // Compute max_diffs for the lower row now, because that needs the
        // contents of argb for the current row, which we will overwrite with
        // residuals before proceeding with the next row.
        var tmp *uint88 = current_max_diffs;
        current_max_diffs = lower_max_diffs;
        lower_max_diffs = tmp8;
        if (y + 2 < height) {
          MaxDiffsForRow(width, width, argb + (y + 1) * width, lower_max_diffs, used_subtract_green);
        }
      }
#endif
      for x = 0; x < width; {
        mode :=
            (modes[(y >> bits) * tiles_per_row + (x >> bits)] >> 8) & 0xff;
        x_end := x + (1 << bits);
        if (x_end > width) x_end = width;
        GetResidual(width, height, upper_row, current_row, current_max_diffs, mode, x, x_end, y, max_quantization, exact, used_subtract_green, argb + y * width + x);
        x = x_end;
      }
    }
  }
}

// Checks whether 'image' can be subsampled by finding the biggest power of 2
// squares (defined by 'best_bits') of uniform value it is made out of.
func VP8LOptimizeSampling(/* const */ image *uint32, int full_width, int full_height, bits int, int max_bits, best_bits_out *int) {
  width := VP8LSubSampleSize(full_width, bits);
  height := VP8LSubSampleSize(full_height, bits);
  int old_width, x, y, square_size;
  best_bits := bits;
  *best_bits_out = bits;
  // Check rows first.
  while (best_bits < max_bits) {
    new_square_size := 1 << (best_bits + 1 - bits);
    is_good := 1;
    square_size = 1 << (best_bits - bits);
    for y = 0; y + square_size < height; y += new_square_size {
      // Check the first lines of consecutive line groups.
      if (memcmp(&image[y * width], &image[(y + square_size) * width], width * sizeof(*image)) != 0) {
        is_good = 0;
        break;
      }
    }
    if (is_good) {
      best_bits++
    } else {
      break;
    }
  }
  if (best_bits == bits) return;

  // Check columns.
  while (best_bits > bits) {
    is_good := 1;
    square_size = 1 << (best_bits - bits);
    for y = 0; is_good && y < height; y++ {
      for x = 0; is_good && x < width; x += square_size {
        var i int
        for i = x + 1; i < GetMin(x + square_size, width); i++ {
          if (image[y * width + i] != image[y * width + x]) {
            is_good = 0;
            break;
          }
        }
      }
    }
    if (is_good) {
      break;
    }
    --best_bits;
  }
  if (best_bits == bits) return;

  // Subsample the image.
  old_width = width;
  square_size = 1 << (best_bits - bits);
  width = VP8LSubSampleSize(full_width, best_bits);
  height = VP8LSubSampleSize(full_height, best_bits);
  for y = 0; y < height; y++ {
    for x = 0; x < width; x++ {
      image[y * width + x] = image[square_size * (y * old_width + x)];
    }
  }
  *best_bits_out = best_bits;
}

// Computes the best predictor image.
// Finds the best predictors per tile. Once done, finds the best predictor image
// sampling.
// best_bits is set to 0 in case of error.
// The following requires some glossary:
// - a tile is a square of side 2^min_bits pixels.
// - a super-tile of a tile is a square of side 2^bits pixels with bits in
// [min_bits+1, max_bits].
// - the max-tile of a tile is the square of 2^max_bits pixels containing it.
//   If this max-tile crosses the border of an image, it is cropped.
// - tile, super-tiles and max_tile are aligned on powers of 2 in the original
//   image.
// - coordinates for tile, super-tile, max-tile are respectively named
//   tile_x, super_tile_x, max_tile_x at their bit scale.
// - in the max-tile, a tile has local coordinates (local_tile_x, local_tile_y).
// The tiles are processed in the following zigzag order to complete the
// super-tiles as soon as possible:
//   1  2|  5  6
//   3  4|  7  8
// --------------
//   9 10| 13 14
//  11 12| 15 16
// When computing the residuals for a tile, the histogram of the above
// super-tile is updated. If this super-tile is finished, its histogram is used
// to update the histogram of the next super-tile and so on up to the max-tile.
func GetBestPredictorsAndSubSampling(
    width, height int, /*const*/ int min_bits, /*const*/ int max_bits, /*const*/ argb_scratch *uint32, /*const*/ argb *uint32, int max_quantization, exact int, int used_subtract_green, /*const*/ pic *WebPPicture, percent_range int, /*const*/ percent *int, *uint32* const all_modes, best_bits *int, *uint32* best_mode) {
  tiles_per_row := VP8LSubSampleSize(width, min_bits);
  tiles_per_col := VP8LSubSampleSize(height, min_bits);
  var best_cost int64
  var subsampling_index uint32
  max_subsampling_index := max_bits - min_bits;
  // Compute the needed memory size for residual histograms, accumulated
  // residual histograms and predictor histograms.
  num_argb := (max_subsampling_index + 1) * kNumPredModes * HISTO_SIZE;
  num_accumulated_rgb := (max_subsampling_index + 1) * HISTO_SIZE;
  num_predictors := (max_subsampling_index + 1) * kNumPredModes;
  var raw_data *uint32 = (*uint32)WebPSafeCalloc(
      num_argb + num_accumulated_rgb + num_predictors, sizeof(uint32));
  var all_argb *uint32 = raw_data;
  var all_accumulated_argb *uint32 = all_argb + num_argb;
  var all_pred_histos *uint32 = all_accumulated_argb + num_accumulated_rgb;
  max_tile_size := 1 << max_subsampling_index;  // in tile size
  percent_start := *percent;
  // When using the residuals of a tile for its super-tiles, you can either:
  // - use each residual to update the histogram of the super-tile, with a cost
  //   of 4 * (1<<n)^2 increment operations (4 for the number of channels, and
  //   (1<<n)^2 for the number of pixels in the tile)
  // - use the histogram of the tile to update the histogram of the super-tile, //   with a cost of HISTO_SIZE (1024)
  // The first method is therefore faster until n==4. 'update_up_to_index'
  // defines the maximum subsampling_index for which the residuals should be
  // individually added to the super-tile histogram.
  update_up_to_index :=
      GetMax(GetMin(4, max_bits), min_bits) - min_bits;
  // Coordinates in the max-tile in tile units.
  local_tile_x := 0, local_tile_y = 0;
  max_tile_x := 0, max_tile_y = 0;
  tile_x := 0, tile_y = 0;

  *best_bits = 0;
  *best_mode = nil;
  if (raw_data == nil) return;

  while (tile_y < tiles_per_col) {
    ComputeResidualsForTile(width, height, tile_x, tile_y, min_bits, update_up_to_index, all_argb, argb_scratch, argb, max_quantization, exact, used_subtract_green);

    // Update all the super-tiles that are complete.
    subsampling_index = 0;
    for {
      super_tile_x := tile_x >> subsampling_index;
      super_tile_y := tile_y >> subsampling_index;
      super_tiles_per_row :=
          VP8LSubSampleSize(width, min_bits + subsampling_index);
      GetBestPredictorForTile(all_argb, subsampling_index, super_tile_x, super_tile_y, super_tiles_per_row, all_accumulated_argb, all_modes, all_pred_histos);
      if (subsampling_index == max_subsampling_index) break;

      // Update the following super-tile histogram if it has not been updated
      // yet.
      subsampling_index++
      if (subsampling_index > update_up_to_index &&
          subsampling_index <= max_subsampling_index) {
        VP8LAddVectorEq(
            GetHistoArgbConst(all_argb, subsampling_index - 1, /*mode=*/0), GetHistoArgb(all_argb, subsampling_index, /*mode=*/0), HISTO_SIZE * kNumPredModes);
      }
      // Check whether the super-tile is not complete (if the smallest tile
      // is not at the end of a line/column or at the beginning of a super-tile
      // of size (1 << subsampling_index)).
      if (!((tile_x == (tiles_per_row - 1) ||
             (local_tile_x + 1) % (1 << subsampling_index) == 0) &&
            (tile_y == (tiles_per_col - 1) ||
             (local_tile_y + 1) % (1 << subsampling_index) == 0))) {
        --subsampling_index;
        // subsampling_index now is the index of the last finished super-tile.
        break;
      }
    }
    // Reset all the histograms belonging to finished tiles.
    stdlib.Memset(all_argb, 0, HISTO_SIZE * kNumPredModes * (subsampling_index + 1) *
               sizeof(*all_argb));

    if (subsampling_index == max_subsampling_index) {
      // If a new max-tile is started.
      if (tile_x == (tiles_per_row - 1)) {
        max_tile_x = 0;
        max_tile_y++
      } else {
        max_tile_x++
      }
      local_tile_x = 0;
      local_tile_y = 0;
    } else {
      // Proceed with the Z traversal.
      coord_x := local_tile_x >> subsampling_index;
      coord_y := local_tile_y >> subsampling_index;
      if (tile_x == (tiles_per_row - 1) && coord_x % 2 == 0) {
        coord_y++
      } else {
        if (coord_x % 2 == 0) {
          coord_x++
        } else {
          // Z traversal.
          coord_y++
          --coord_x;
        }
      }
      local_tile_x = coord_x << subsampling_index;
      local_tile_y = coord_y << subsampling_index;
    }
    tile_x = max_tile_x * max_tile_size + local_tile_x;
    tile_y = max_tile_y * max_tile_size + local_tile_y;

    if (tile_x == 0 &&
        !WebPReportProgress(
            pic, percent_start + percent_range * tile_y / tiles_per_col, percent)) {
      return;
    }
  }

  // Figure out the best sampling.
  best_cost = WEBP_INT64_MAX;
  for (subsampling_index = 0; subsampling_index <= max_subsampling_index;
       ++subsampling_index) {
    var plane int
    const accumulated *uint32 =
        GetAccumulatedHisto(all_accumulated_argb, subsampling_index);
    cost := VP8LShannonEntropy(
        &all_pred_histos[subsampling_index * kNumPredModes], kNumPredModes);
    for plane = 0; plane < 4; plane++ {
      cost += VP8LShannonEntropy(&accumulated[plane * 256], 256);
    }
    if (cost < best_cost) {
      best_cost = cost;
      *best_bits = min_bits + subsampling_index;
      *best_mode = all_modes[subsampling_index];
    }
  }

  VP8LOptimizeSampling(*best_mode, width, height, *best_bits, MAX_TRANSFORM_BITS, best_bits);
}

// Finds the best predictor for each tile, and converts the image to residuals
// with respect to predictions. If near_lossless_quality < 100, applies
// near lossless processing, shaving off more bits of residuals for lower
// qualities.
// pic and percent are for progress.
// Returns false in case of error (stored in pic.error_code).
int VP8LResidualImage(width, height int, int min_bits, int max_bits, low_effort int, /*const*/ argb *uint32, /*const*/ argb_scratch *uint32, /*const*/ image *uint32, int near_lossless_quality, exact int, int used_subtract_green, /*const*/ pic *WebPPicture, percent_range int, /*const*/ percent *int, /*const*/ best_bits *int) {
  percent_start := *percent;
  max_quantization := 1 << VP8LNearLosslessBits(near_lossless_quality);
  if (low_effort) {
    tiles_per_row := VP8LSubSampleSize(width, max_bits);
    tiles_per_col := VP8LSubSampleSize(height, max_bits);
    var i int
    for i = 0; i < tiles_per_row * tiles_per_col; i++ {
      image[i] = ARGB_BLACK | (kPredLowEffort << 8);
    }
    *best_bits = max_bits;
  } else {
    // Allocate data to try all samplings from min_bits to max_bits.
    bits int;
    sum_num_pixels := uint(0);
    uint32 *modes_raw, *best_mode;
    modes *uint32[MAX_TRANSFORM_BITS + 1];
    uint32 num_pixels[MAX_TRANSFORM_BITS + 1];
    for bits = min_bits; bits <= max_bits; bits++ {
      tiles_per_row := VP8LSubSampleSize(width, bits);
      tiles_per_col := VP8LSubSampleSize(height, bits);
      num_pixels[bits] = tiles_per_row * tiles_per_col;
      sum_num_pixels += num_pixels[bits];
    }
    modes_raw = (*uint32)WebPSafeMalloc(sum_num_pixels, sizeof(*modes_raw));
    if (modes_raw == nil) { return 0; }
    // Have modes point to the right global memory modes_raw.
    modes[min_bits] = modes_raw;
    for bits = min_bits + 1; bits <= max_bits; bits++ {
      modes[bits] = modes[bits - 1] + num_pixels[bits - 1];
    }
    // Find the best sampling.
    GetBestPredictorsAndSubSampling(
        width, height, min_bits, max_bits, argb_scratch, argb, max_quantization, exact, used_subtract_green, pic, percent_range, percent, &modes[min_bits], best_bits, &best_mode);
    if (*best_bits == 0) {
      return 0;
    }
    // Keep the best predictor image.
    memcpy(image, best_mode, VP8LSubSampleSize(width, *best_bits) *
               VP8LSubSampleSize(height, *best_bits) * sizeof(*image));
  }

  CopyImageWithPrediction(width, height, *best_bits, image, argb_scratch, argb, low_effort, max_quantization, exact, used_subtract_green);
  return WebPReportProgress(pic, percent_start + percent_range, percent);
}

//------------------------------------------------------------------------------
// Color transform functions.

static  func MultipliersClear(/* const */ m *VP8LMultipliers) {
  m.green_to_red = 0;
  m.green_to_blue = 0;
  m.red_to_blue = 0;
}

static  func ColorCodeToMultipliers(uint32 color_code, /*const*/ m *VP8LMultipliers) {
  m.green_to_red = (color_code >> 0) & 0xff;
  m.green_to_blue = (color_code >> 8) & 0xff;
  m.red_to_blue = (color_code >> 16) & 0xff;
}

static  uint32
MultipliersToColorCode(/* const */ m *VP8LMultipliers) {
  return uint(0xff000000) | ((uint32)(m.red_to_blue) << 16) |
         ((uint32)(m.green_to_blue) << 8) | m.green_to_red;
}

func PredictionCostCrossColor(/* const */ uint32 accumulated[256], /*const*/ uint32 counts[256]) int64 {
  // Favor low entropy, locally and globally.
  // Favor small absolute values for PredictionCostSpatial
  static const kExpValue := 240;
  return (int64)VP8LCombinedShannonEntropy(counts, accumulated) +
         PredictionCostBias(counts, 3, kExpValue);
}

static int64 GetPredictionCostCrossColorRed(
    const argb *uint32, int stride, int tile_width, int tile_height, VP8LMultipliers prev_x, VP8LMultipliers prev_y, int green_to_red, /*const*/ uint32 accumulated_red_histo[256]) {
  uint32 histo[256] = {0}
  var cur_diff int64

  VP8LCollectColorRedTransforms(argb, stride, tile_width, tile_height, green_to_red, histo);

  cur_diff = PredictionCostCrossColor(accumulated_red_histo, histo);
  if ((uint8)green_to_red == prev_x.green_to_red) {
    // favor keeping the areas locally similar
    cur_diff -= 3ll << LOG_2_PRECISION_BITS;
  }
  if ((uint8)green_to_red == prev_y.green_to_red) {
    // favor keeping the areas locally similar
    cur_diff -= 3ll << LOG_2_PRECISION_BITS;
  }
  if (green_to_red == 0) {
    cur_diff -= 3ll << LOG_2_PRECISION_BITS;
  }
  return cur_diff;
}

func GetBestGreenToRed(/* const */ argb *uint32, int stride, int tile_width, int tile_height, VP8LMultipliers prev_x, VP8LMultipliers prev_y, quality int, /*const*/ uint32 accumulated_red_histo[256], /*const*/ best_tx *VP8LMultipliers) {
  kMaxIters := 4 + ((7 * quality) >> 8);  // in range [4..6]
  green_to_red_best := 0;
  int iter, offset;
  best_diff := GetPredictionCostCrossColorRed(
      argb, stride, tile_width, tile_height, prev_x, prev_y, green_to_red_best, accumulated_red_histo);
  for iter = 0; iter < kMaxIters; iter++ {
    // ColorTransformDelta is a 3.5 bit fixed point, so 32 is equal to
    // one in color computation. Having initial delta here as 1 is sufficient
    // to explore the range of (-2, 2).
    delta := 32 >> iter;
    // Try a negative and a positive delta from the best known value.
    for offset = -delta; offset <= delta; offset += 2 * delta {
      green_to_red_cur := offset + green_to_red_best;
      cur_diff := GetPredictionCostCrossColorRed(
          argb, stride, tile_width, tile_height, prev_x, prev_y, green_to_red_cur, accumulated_red_histo);
      if (cur_diff < best_diff) {
        best_diff = cur_diff;
        green_to_red_best = green_to_red_cur;
      }
    }
  }
  best_tx.green_to_red = (green_to_red_best & 0xff);
}

static int64 GetPredictionCostCrossColorBlue(
    const argb *uint32, int stride, int tile_width, int tile_height, VP8LMultipliers prev_x, VP8LMultipliers prev_y, int green_to_blue, int red_to_blue, /*const*/ uint32 accumulated_blue_histo[256]) {
  uint32 histo[256] = {0}
  var cur_diff int64

  VP8LCollectColorBlueTransforms(argb, stride, tile_width, tile_height, green_to_blue, red_to_blue, histo);

  cur_diff = PredictionCostCrossColor(accumulated_blue_histo, histo);
  if ((uint8)green_to_blue == prev_x.green_to_blue) {
    // favor keeping the areas locally similar
    cur_diff -= 3ll << LOG_2_PRECISION_BITS;
  }
  if ((uint8)green_to_blue == prev_y.green_to_blue) {
    // favor keeping the areas locally similar
    cur_diff -= 3ll << LOG_2_PRECISION_BITS;
  }
  if ((uint8)red_to_blue == prev_x.red_to_blue) {
    // favor keeping the areas locally similar
    cur_diff -= 3ll << LOG_2_PRECISION_BITS;
  }
  if ((uint8)red_to_blue == prev_y.red_to_blue) {
    // favor keeping the areas locally similar
    cur_diff -= 3ll << LOG_2_PRECISION_BITS;
  }
  if (green_to_blue == 0) {
    cur_diff -= 3ll << LOG_2_PRECISION_BITS;
  }
  if (red_to_blue == 0) {
    cur_diff -= 3ll << LOG_2_PRECISION_BITS;
  }
  return cur_diff;
}

const kGreenRedToBlueNumAxis = 8
const kGreenRedToBlueMaxIters = 7
func GetBestGreenRedToBlue(/* const */ argb *uint32, int stride, int tile_width, int tile_height, VP8LMultipliers prev_x, VP8LMultipliers prev_y, quality int, /*const*/ uint32 accumulated_blue_histo[256], /*const*/ best_tx *VP8LMultipliers) {
  offset[kGreenRedToBlueNumAxis][2] := {
      {0, -1}, {0, 1}, {-1, 0}, {1, 0}, {-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
  delta_lut[kGreenRedToBlueMaxIters] := {16, 16, 8, 4, 2, 2, 2}
  iters := (quality < 25)   ? 1
                    : (quality > 50) ? kGreenRedToBlueMaxIters
                                     : 4;
  green_to_blue_best := 0;
  red_to_blue_best := 0;
  var iter int
  // Initial value at origin:
  best_diff := GetPredictionCostCrossColorBlue(
      argb, stride, tile_width, tile_height, prev_x, prev_y, green_to_blue_best, red_to_blue_best, accumulated_blue_histo);
  for iter = 0; iter < iters; iter++ {
    delta := delta_lut[iter];
    var axis int
    for axis = 0; axis < kGreenRedToBlueNumAxis; axis++ {
      green_to_blue_cur :=
          offset[axis][0] * delta + green_to_blue_best;
      red_to_blue_cur := offset[axis][1] * delta + red_to_blue_best;
      cur_diff := GetPredictionCostCrossColorBlue(
          argb, stride, tile_width, tile_height, prev_x, prev_y, green_to_blue_cur, red_to_blue_cur, accumulated_blue_histo);
      if (cur_diff < best_diff) {
        best_diff = cur_diff;
        green_to_blue_best = green_to_blue_cur;
        red_to_blue_best = red_to_blue_cur;
      }
      if (quality < 25 && iter == 4) {
        // Only axis aligned diffs for lower quality.
        break;  // next iter.
      }
    }
    if (delta == 2 && green_to_blue_best == 0 && red_to_blue_best == 0) {
      // Further iterations would not help.
      break;  // out of iter-loop.
    }
  }
  best_tx.green_to_blue = green_to_blue_best & 0xff;
  best_tx.red_to_blue = red_to_blue_best & 0xff;
}
#undef kGreenRedToBlueMaxIters
#undef kGreenRedToBlueNumAxis

static VP8LMultipliers GetBestColorTransformForTile(
    int tile_x, int tile_y, bits int, VP8LMultipliers prev_x, VP8LMultipliers prev_y, quality int, int xsize, int ysize, /*const*/ uint32 accumulated_red_histo[256], /*const*/ uint32 accumulated_blue_histo[256], /*const*/ argb *uint32) {
  max_tile_size := 1 << bits;
  tile_y_offset := tile_y * max_tile_size;
  tile_x_offset := tile_x * max_tile_size;
  all_x_max := GetMin(tile_x_offset + max_tile_size, xsize);
  all_y_max := GetMin(tile_y_offset + max_tile_size, ysize);
  tile_width := all_x_max - tile_x_offset;
  tile_height := all_y_max - tile_y_offset;
  const tile_argb *uint32 =
      argb + tile_y_offset * xsize + tile_x_offset;
  VP8LMultipliers best_tx;
  MultipliersClear(&best_tx);

  GetBestGreenToRed(tile_argb, xsize, tile_width, tile_height, prev_x, prev_y, quality, accumulated_red_histo, &best_tx);
  GetBestGreenRedToBlue(tile_argb, xsize, tile_width, tile_height, prev_x, prev_y, quality, accumulated_blue_histo, &best_tx);
  return best_tx;
}

func CopyTileWithColorTransform(int xsize, int ysize, int tile_x, int tile_y, int max_tile_size, VP8LMultipliers color_transform, argb *uint32) {
  xscan := GetMin(max_tile_size, xsize - tile_x);
  yscan := GetMin(max_tile_size, ysize - tile_y);
  argb += tile_y * xsize + tile_x;
  while (yscan-- > 0) {
    VP8LTransformColor(&color_transform, argb, xscan);
    argb += xsize;
  }
}

int VP8LColorSpaceTransform(width, height int, bits int, quality int, /*const*/ argb *uint32, image *uint32, /*const*/ pic *WebPPicture, percent_range int, /*const*/ percent *int, /*const*/ best_bits *int) {
  max_tile_size := 1 << bits;
  tile_xsize := VP8LSubSampleSize(width, bits);
  tile_ysize := VP8LSubSampleSize(height, bits);
  percent_start := *percent;
  uint32 accumulated_red_histo[256] = {0}
  uint32 accumulated_blue_histo[256] = {0}
  int tile_x, tile_y;
  VP8LMultipliers prev_x, prev_y;
  MultipliersClear(&prev_y);
  MultipliersClear(&prev_x);
  for tile_y = 0; tile_y < tile_ysize; tile_y++ {
    for tile_x = 0; tile_x < tile_xsize; tile_x++ {
      var y int
      tile_x_offset := tile_x * max_tile_size;
      tile_y_offset := tile_y * max_tile_size;
      all_x_max := GetMin(tile_x_offset + max_tile_size, width);
      all_y_max := GetMin(tile_y_offset + max_tile_size, height);
      offset := tile_y * tile_xsize + tile_x;
      if (tile_y != 0) {
        ColorCodeToMultipliers(image[offset - tile_xsize], &prev_y);
      }
      prev_x = GetBestColorTransformForTile(
          tile_x, tile_y, bits, prev_x, prev_y, quality, width, height, accumulated_red_histo, accumulated_blue_histo, argb);
      image[offset] = MultipliersToColorCode(&prev_x);
      CopyTileWithColorTransform(width, height, tile_x_offset, tile_y_offset, max_tile_size, prev_x, argb);

      // Gather accumulated histogram data.
      for y = tile_y_offset; y < all_y_max; y++ {
        ix := y * width + tile_x_offset;
        ix_end := ix + all_x_max - tile_x_offset;
        for ; ix < ix_end; ix++ {
          pix := argb[ix];
          if (ix >= 2 && pix == argb[ix - 2] && pix == argb[ix - 1]) {
            continue;  // repeated pixels are handled by backward references
          }
          if (ix >= width + 2 && argb[ix - 2] == argb[ix - width - 2] &&
              argb[ix - 1] == argb[ix - width - 1] && pix == argb[ix - width]) {
            continue;  // repeated pixels are handled by backward references
          }
          ++accumulated_red_histo[(pix >> 16) & 0xff];
          ++accumulated_blue_histo[(pix >> 0) & 0xff];
        }
      }
    }
    if (!WebPReportProgress(pic, percent_start + percent_range * tile_y / tile_ysize, percent)) {
      return 0;
    }
  }
  VP8LOptimizeSampling(image, width, height, bits, MAX_TRANSFORM_BITS, best_bits);
  return 1;
}
