// Copyright 2014 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package dsp

import (
	"github.com/daanv2/go-webp/pkg/assert"
	"github.com/daanv2/go-webp/pkg/libwebp/enc"
	"github.com/daanv2/go-webp/pkg/stdlib"
	"github.com/daanv2/go-webp/pkg/util/tenary"
)

func VP8GetResidualCost(ctx0 int, res *enc.VP8Residual) int {
	n := res.first
	// should be prob[VP8EncBands[n]], but it's equivalent for n=0 or 1
	p0 := res.prob[n][ctx0][0]
	costs := res.costs
	var t *uint16 = costs[n][ctx0]
	// bit_cost(1, p0) is already incorporated in t[] tables, but only if ctx != 0
	// (as required by the syntax). For ctx0 == 0, we need to add it here or it'll
	// be missing during the loop.
	cost := tenary.If((ctx0 == 0), enc.VP8BitCost(1, p0), 0)

	if res.last < 0 {
		return enc.VP8BitCost(0, p0)
	}
	for ; n < res.last; n++ {
		v := stdlib.Abs(res.coeffs[n])
		ctx := tenary.If((v >= 2), 2, v)
		cost += enc.VP8LevelCost(t, v)
		t = costs[n+1][ctx]
	}
	// Last coefficient is always non-zero
	{
		v := stdlib.Abs(res.coeffs[n])
		assert.Assert(v != 0)
		cost += enc.VP8LevelCost(t, v)
		if n < 15 {
			b := VP8EncBands[n+1]
			ctx := tenary.If((v == 1), 1, 2)
			last_p0 := res.prob[b][ctx][0]
			cost += enc.VP8BitCost(0, last_p0)
		}
	}
	return cost
}

func VP8SetResidualCoeffs( /* const */ /* const */ coeffs []int16 /* const */, res *enc.VP8Residual) {
	var n int
	res.last = -1
	assert.Assert(res.first == 0 || coeffs[0] == 0)
	for n = 15; n >= 0; n-- {
		if coeffs[n] != 0 {
			res.last = n
			break
		}
	}
	res.coeffs = coeffs
}
