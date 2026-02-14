// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package vp8

// These five modes are evaluated and their respective entropy is computed.
type EntropyIx int

const (
	kDirect            EntropyIx = 0
	kSpatial           EntropyIx = 1
	kSubGreen          EntropyIx = 2
	kSpatialSubGreen   EntropyIx = 3
	kPalette           EntropyIx = 4
	kPaletteAndSpatial EntropyIx = 5
	kNumEntropyIx      EntropyIx = 6
)


