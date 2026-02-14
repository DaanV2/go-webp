// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package dsp

type WEBP_FILTER_TYPE int

const ( // Filter types.
	WEBP_FILTER_NONE WEBP_FILTER_TYPE = iota
	WEBP_FILTER_HORIZONTAL
	WEBP_FILTER_VERTICAL
	WEBP_FILTER_GRADIENT
	WEBP_FILTER_BEST // meta-types
	WEBP_FILTER_FAST
)

const (
	WEBP_FILTER_LAST = WEBP_FILTER_GRADIENT + 1 // end marker
)
