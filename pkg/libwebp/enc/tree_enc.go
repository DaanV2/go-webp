// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package enc

import (
	"github.com/daanv2/go-webp/pkg/libwebp/decoder"
	"github.com/daanv2/go-webp/pkg/stdlib"
	"github.com/daanv2/go-webp/pkg/vp8"
)

// Reset the token probabilities to their initial (default) values
func VP8DefaultProbas( /* const */ enc *vp8.VP8Encoder) {
	var probas *vp8.VP8EncProba = &enc.proba
	probas.use_skip_proba = 0
	stdlib.Memset(probas.segments, uint(255), sizeof(probas.segments))
	stdlib.MemCpy(probas.coeffs, VP8CoeffsProba0, sizeof(VP8CoeffsProba0))
	// Note: we could hard-code the level_costs corresponding to VP8CoeffsProba0, // but that's ~11k of static data. Better call VP8CalculateLevelCosts() later.
	probas.dirty = 1
}

func PutI4Mode( /* const */ bw *vp8.VP8BitWriter, mode decoder.PREDICTION_MODE /*const*/, prob []uint8) decoder.PREDICTION_MODE {
	if vp8.VP8PutBit(bw, mode != decoder.B_DC_PRED, int(prob[0])) {
		if vp8.VP8PutBit(bw, mode != decoder.B_TM_PRED, int(prob[1])) {
			if vp8.VP8PutBit(bw, mode != decoder.B_VE_PRED, int(prob[2])) {
				if !vp8.VP8PutBit(bw, mode >= decoder.B_LD_PRED, int(prob[3])) {
					if vp8.VP8PutBit(bw, mode != decoder.B_HE_PRED, int(prob[4])) {
						vp8.VP8PutBit(bw, mode != decoder.B_RD_PRED, int(prob[5]))
					}
				} else {
					if vp8.VP8PutBit(bw, mode != decoder.B_LD_PRED, int(prob[6])) {
						if vp8.VP8PutBit(bw, mode != decoder.B_VL_PRED, int(prob[7])) {
							vp8.VP8PutBit(bw, mode != decoder.B_HD_PRED, int(prob[8]))
						}
					}
				}
			}
		}
	}
	return mode
}

func PutI16Mode( /* const */ bw *vp8.VP8BitWriter, mode decoder.PREDICTION_MODE) {
	if vp8.VP8PutBit(bw, (mode == decoder.TM_PRED || mode == decoder.H_PRED), 156) {
		vp8.VP8PutBit(bw, mode == decoder.TM_PRED, 128) // TM or HE
	} else {
		vp8.VP8PutBit(bw, mode == decoder.V_PRED, 163) // VE or DC
	}
}

func PutUVMode( /* const */ bw *vp8.VP8BitWriter, uv_mode decoder.PREDICTION_MODE) {
	if vp8.VP8PutBit(bw, uv_mode != decoder.DC_PRED, 142) {
		if vp8.VP8PutBit(bw, uv_mode != decoder.V_PRED, 114) {
			vp8.VP8PutBit(bw, uv_mode != decoder.H_PRED, 183) // else: TM_PRED
		}
	}
}

func PutSegment( /* const */ bw *vp8.VP8BitWriter, s int /*const*/, p []uint8) {
	if vp8.VP8PutBit(bw, s >= 2, int(p[0])) {
		p = p[1:]
	}
	vp8.VP8PutBit(bw, (s&1) != 0, int(p[1]))
}

// Writes the partition #0 modes (that is: all intra modes)
func VP8CodeIntraModes( /* const */ enc *vp8.VP8Encoder) {
	var bw *vp8.VP8BitWriter = &enc.bw
	var it vp8.VP8EncIterator
	VP8IteratorInit(enc, &it)
	for {
		var mb *vp8.VP8MBInfo = it.mb
		var preds uint8 = it.preds
		if enc.segment_hdr.update_map {
			PutSegment(bw, mb.segment, enc.proba.segments)
		}
		if enc.proba.use_skip_proba {
			vp8.VP8PutBit(bw, mb.skip, enc.proba.skip_proba)
		}
		if vp8.VP8PutBit(bw, (mb.vtype != 0), 145) { // i16x16
			PutI16Mode(bw, preds[0])
		} else {
			preds_w := enc.preds_w
			var top_pred *uint8 = preds - preds_w
			var x, y int
			for y = 0; y < 4; y++ {
				left := preds[-1]
				for x = 0; x < 4; x++ {
					probas := kBModesProba[top_pred[x]][left]
					left = PutI4Mode(bw, preds[x], probas[:])
				}
				top_pred = preds
				preds += preds_w
			}
		}
		PutUVMode(bw, mb.uv_mode)
		if VP8IteratorNext(&it) == 0 {
			break
		}
	}
}

// Write the token probabilities
func VP8WriteProbas( /* const */ bw *vp8.VP8BitWriter /*const*/, probas *vp8.VP8EncProba) {
	var t, b, c, p int

	for t = 0; t < vp8.NUM_TYPES; t++ {
		for b = 0; b < vp8.NUM_BANDS; b++ {
			for c = 0; c < vp8.NUM_CTX; c++ {
				for p = 0; p < vp8.NUM_PROBAS; p++ {
					p0 := probas.coeffs[t][b][c][p]
					update := (p0 != VP8CoeffsProba0[t][b][c][p])
					if vp8.VP8PutBit(bw, update, int(VP8CoeffsUpdateProba[t][b][c][p])) {
						vp8.VP8PutBits(bw, p0, 8)
					}
				}
			}
		}
	}
	if vp8.VP8PutBitUniform(bw, probas.use_skip_proba != 0) {
		vp8.VP8PutBits(bw, probas.skip_proba, 8)
	}
}
