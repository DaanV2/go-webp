package vp8

// Copyright 2010 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
// -----------------------------------------------------------------------------
//
//  Low-level API for VP8 decoder
//
// Author: Skal (pascal.massimino@gmail.com)


import "github.com/daanv2/go-webp/pkg/stddef"

import "github.com/daanv2/go-webp/pkg/libwebp/webp"
import "github.com/daanv2/go-webp/pkg/libwebp/webp"


//------------------------------------------------------------------------------
// Lower-level API
//
// These functions provide fine-grained control of the decoding process.
// The call flow should resemble:
//
//   VP8Io io;
//   VP8InitIo(&io);
//   io.data = data;
//   io.data_size = size;
//   /* customize io's functions (setup()/put()/teardown()) if needed. */
//
//   dec *VP8Decoder = VP8New();
//   ok := VP8Decode(dec, &io);
//   if (!ok) printf("Error: %s\n", VP8StatusMessage(dec));
//   VP8Delete(dec);
//   return ok;



// Internal, version-checked, entry point
 int VP8InitIoInternal(const *VP8Io, int);

// Set the custom IO function pointers and user-data. The setter for IO hooks
// should be called before initiating incremental decoding. Returns true if
// WebPIDecoder object is successfully modified, false otherwise.
 int WebPISetIOHooks(const idec *WebPIDecoder, VP8IoPutHook put, VP8IoSetupHook setup, VP8IoTeardownHook teardown, user_data *void);

// Main decoding object. This is an opaque structure.
typedef struct VP8Decoder VP8Decoder;

// Create a new decoder object.
VP *VP8Decoder8New(void);

// Must be called to make sure 'io' is initialized properly.
// Returns false in case of version mismatch. Upon such failure, no other
// decoding function should be called (VP8Decode, VP8GetHeaders, ...)
 static  int VP8InitIo(const io *VP8Io) {
  return VP8InitIoInternal(io, WEBP_DECODER_ABI_VERSION);
}

// Decode the VP8 frame header. Returns true if ok.
// Note: 'io.data' must be pointing to the start of the VP8 frame header.
 int VP8GetHeaders(const dec *VP8Decoder, const io *VP8Io);

// Decode a picture. Will call VP8GetHeaders() if it wasn't done already.
// Returns false in case of error.
 int VP8Decode(const dec *VP8Decoder, const io *VP8Io);

// Return current status of the decoder:
VP8StatusCode VP8Status(const dec *VP8Decoder);

// return readable string corresponding to the last status.
const VP *byte8StatusMessage(const dec *VP8Decoder);

// Resets the decoder in its initial state, reclaiming memory.
// Not a mandatory call between calls to VP8Decode().
func VP8Clear(const dec *VP8Decoder);

// Destroy the decoder object.
func VP8Delete(const dec *VP8Decoder);

//------------------------------------------------------------------------------
// Miscellaneous VP8/VP8L bitstream probing functions.

// Returns true if the next 3 bytes in data contain the VP8 signature.
 int VP8CheckSignature(
    const *uint8  data, uint64 data_size);

// Validates the VP8 data-header and retrieves basic header information viz
// width and height. Returns 0 in case of formatting error. *width/*height
// can be passed nil.
 int VP8GetInfo(
    const *uint8  data, uint64 data_size,   // data available so far
    uint64 chunk_size,  // total data size expected in the chunk
    const width *int, const height *int);

// Returns true if the next byte(s) in data is a VP8L signature.
 int VP8LCheckSignature(const *uint8 
                                       data, size uint64 );

// Validates the VP8L data-header and retrieves basic header information viz
// width, height and alpha. Returns 0 in case of formatting error.
// width/height/has_alpha can be passed nil.
 int VP8LGetInfo(const *uint8  data, uint64 data_size,  // data available so far
                            const width *int, const height *int, const has_alpha *int);



#endif  // WEBP_DEC_VP8_DEC_H_
