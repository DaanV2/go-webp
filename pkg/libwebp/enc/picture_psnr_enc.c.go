// Copyright 2014 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
package enc

import "github.com/daanv2/go-webp/pkg/libwebp/webp"

func WebPPlaneDistortion(/* const */ src *uint8, uint64 src_stride, /*const*/ ref *uint8, uint64 ref_stride, width, height int, uint64 x_step, vtype int, distortion *float64, result *float64) int {
//   (void)src
//   (void)src_stride
//   (void)ref
//   (void)ref_stride
//   (void)width
//   (void)height
//   (void)x_step
//   (void)vtype
  if distortion == nil || result == nil { return 0  }
  *distortion = 0.0
  *result = 0.0
  return 1
}

func WebPPictureDistortion(/* const */ src *picture.Picture, /*const*/ ref *picture.Picture, vtype int, results [5]float64) int {
  var i int
//   (void)src
//   (void)ref
//   (void)type
  if results == nil { return 0  }
  for (i = 0; i < 5; ++i) results[i] = 0.0
  return 1
}
