// Copyright 2014 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package mux

// Stores frame rectangle dimensions.
type FrameRectangle struct {
   x_offset, y_offset, width, height int
}

// Used to store two candidates of encoded data for an animation frame. One of
// the two will be chosen later.
type EncodedFrame struct {
   sub_frame WebPMuxFrameInfo  // Encoded frame rectangle.
   key_frame WebPMuxFrameInfo  // Encoded frame if it is a keyframe.
   is_key_frame int            // True if 'key_frame' has been chosen.
}