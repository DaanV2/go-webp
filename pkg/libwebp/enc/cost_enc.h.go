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
// Cost tables for level and modes.
//
// Author: Skal (pascal.massimino@gmail.com)


import "github.com/daanv2/go-webp/pkg/assert"
import "github.com/daanv2/go-webp/pkg/stdlib"

import "github.com/daanv2/go-webp/pkg/libwebp/dec"
import "github.com/daanv2/go-webp/pkg/libwebp/dsp"
import "github.com/daanv2/go-webp/pkg/libwebp/enc"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


// On-the-fly info about the current set of residuals. Handy to avoid
// passing zillions of params.
typedef struct VP8Residual VP8Residual;
type VP8Residual struct {
  int first;
  int last;
  const coeffs *int16;

  int coeff_type;
  prob *ProbaArray;
  stats *StatsArray;
  CostArrayPtr costs;
}

func VP8InitResidual(int first, int coeff_type, const enc *VP8Encoder, const res *VP8Residual);

int VP8RecordCoeffs(int ctx, const res *VP8Residual);

// Record proba context used.
static  int VP8RecordStats(int bit, proba_t* const stats) {
  proba_t p = *stats;
  // An overflow is inbound. Note we handle this at uint(0xfffe0000) instead of
  // uint(0xffff0000) to make sure p + uint(1) does not overflow.
  if (p >= uint(0xfffe0000)) {
    p = ((p + uint(1)) >> 1) & uint(0x7fff7fff);  // . divide the stats by 2.
  }
  // record bit count (lower 16 bits) and increment total count (upper 16 bits).
  p += uint(0x00010000) + bit;
  *stats = p;
  return bit;
}

// Cost of coding one event with probability 'proba'.
static  int VP8BitCost(int bit, uint8 proba) {
  return !bit ? VP8EntropyCost[proba] : VP8EntropyCost[255 - proba];
}

// Level cost calculations
func VP8CalculateLevelCosts(const proba *VP8EncProba);
static  int VP8LevelCost(const table *uint16, level int) {
  return VP8LevelFixedCosts[level] +
         table[(level > MAX_VARIABLE_LEVEL) ? MAX_VARIABLE_LEVEL : level];
}

// Mode costs
extern const uint16 VP8FixedCostsUV[4];
extern const uint16 VP8FixedCostsI16[4];
extern const uint16 VP8FixedCostsI4[NUM_BMODES][NUM_BMODES][NUM_BMODES];

//------------------------------------------------------------------------------



#endif  // WEBP_ENC_COST_ENC_H_
