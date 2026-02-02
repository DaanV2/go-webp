// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package huffman

func SetBitDepths(tree *HuffmanTree, pool []*HuffmanTree, bit_depths []uint8, level int) {
	if tree.pool_index_left >= 0 {
		SetBitDepths(pool[tree.pool_index_left], pool, bit_depths, level+1)
		SetBitDepths(pool[tree.pool_index_right], pool, bit_depths, level+1)
	} else {
		bit_depths[tree.value] = uint8(level)
	}
}
