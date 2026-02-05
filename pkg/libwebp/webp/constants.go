// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package webp

import "github.com/daanv2/go-webp/pkg/constants"

const WEBP_DECODER_ABI_VERSION = constants.WEBP_DECODER_ABI_VERSION // MAJOR(8b) + MINOR(8b)

// Return the decoder's version number, packed in hexadecimal using 8bits for
// each of major/minor/revision. E.g: v2.5.7 is 0x020507.
func WebPGetDecoderVersion() int {
	return dec.WebPGetDecoderVersion()
}
