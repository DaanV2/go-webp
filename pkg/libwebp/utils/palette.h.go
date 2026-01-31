package utils

// Copyright 2023 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
// Utilities for palette analysis.
//
// Author: Vincent Rabaud (vrabaud@google.com)

// The different ways a palette can be sorted.
type PaletteSorting int

const(
	// Sorts by minimizing L1 deltas between consecutive colors, giving more
	// weight to RGB colors.
	kSortedDefault PaletteSorting = 0
	// Implements the modified Zeng method from "A Survey on Palette Reordering
	// Methods for Improving the Compression of Color-Indexed Images" by Armando
	// J. Pinho and Antonio J. R. Neves.
	kMinimizeDelta PaletteSorting = 1

	kModifiedZeng PaletteSorting = 2
	kUnusedPalette PaletteSorting = 3
	kPaletteSortingNum PaletteSorting = 4
)

// Returns the index of 'color' in the sorted palette 'sorted' of size
// 'num_colors'.
func SearchColorNoIdx(  sorted []uint32, color uint32 , num_colors int ) int;

// Sort palette in increasing order and prepare an inverse mapping array.
func PrepareMapToPalette(palette []uint32,  num_colors uint32, sorted []uint32, idx_map []uint32);

// Returns count of unique colors in 'pic', assuming pic.use_argb is true.
// If the unique color count is more than MAX_PALETTE_SIZE, returns
// MAX_PALETTE_SIZE+1.
// If 'palette' is not nil and the number of unique colors is less than or
// equal to MAX_PALETTE_SIZE, also outputs the actual unique colors into
// 'palette' in a sorted order. Note: 'palette' is assumed to be an array
// already allocated with at least MAX_PALETTE_SIZE elements.
func GetColorPalette(pic *WebPPicture, palette []uint32) int;

// Sorts the palette according to the criterion defined by 'method'.
// 'palette_sorted' is the input palette sorted lexicographically, as done in
// PrepareMapToPalette. Returns 0 on memory allocation error.
// For kSortedDefault and kMinimizeDelta methods, 0 (if present) is set as the
// last element to optimize later storage.
func PaletteSort(method PaletteSorting, pic *WebPPicture, palette_sorted *uint32, num_colors uint32, palette *uint32) int;
