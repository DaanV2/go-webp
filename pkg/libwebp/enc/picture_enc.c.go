package enc

// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

func WebPMemoryWriterInit(writer *WebPMemoryWriter) {
  writer.mem = nil;
  writer.size = 0;
  writer.max_size = 0;
}

// The custom writer to be used with WebPMemoryWriter as custom_ptr. Upon
// completion, writer.mem and writer.size will hold the coded data.
// writer.mem must be freed by calling WebPMemoryWriterClear.
func WebPMemoryWrite(/* const */ data *uint8, data_size uint64, /*const*/ picture *picture.Picture) int {
  var w *WebPMemoryWriter = (*WebPMemoryWriter)picture.CustomPtr;
  var next_size uint64
  if (w == nil) {
    return 1;
  }
  next_size = (uint64)w.size + data_size;
  if (next_size > w.max_size) {
    new_mem *uint8;
    next_max_size := uint64(2) * w.max_size;
    if next_max_size < next_size { next_max_size = next_size }
    if next_max_size < uint64(8192) { next_max_size = uint64(8192) }
    // new_mem = (*uint8)WebPSafeMalloc(next_max_size, 1);
    // if (new_mem == nil) {
    //   return 0;
    // }
	new_mem := make([]uint8, next_max_size)

    if (w.size > 0) {
      stdlib.MemCpy(new_mem, w.mem, w.size);
    }
    w.mem = new_mem;
    // down-cast is ok, thanks to WebPSafeMalloc
    w.max_size = (uint64)next_max_size;
  }
  if (data_size > 0) {
    stdlib.MemCpy(w.mem + w.size, data, data_size);
    w.size += data_size;
  }
  return 1;
}

func WebPMemoryWriterClear(writer *WebPMemoryWriter) {
  if (writer != nil) {
    WebPMemoryWriterInit(writer);
  }
}

//------------------------------------------------------------------------------
// Simplest high-level calls:

type Importer = func(/* const */ *picture.Picture, /*const*/ *uint8, int)int

func Encode(/* const */ rgba *uint8, width, height int, stride int, import Importer, float64 quality_factor, lossless int, out *uint8 put) uint64 {
   var pic picture.Picture
   var config config.Config
   var wrt WebPMemoryWriter
  var ok int

  if output == nil { return 0  }

  if (!WebPConfigPreset(&config, WEBP_PRESET_DEFAULT, quality_factor) ||
      !picture.WebPPictureInit(&pic)) {
    return 0;  // shouldn't happen, except if system installation is broken
  }

  config.Lossless = !!lossless;
  pic.use_argb = !!lossless;
  pic.width = width;
  pic.height = height;
  pic.writer = WebPMemoryWrite;
  pic.custom_ptr = &wrt;
  WebPMemoryWriterInit(&wrt);

  ok = import(&pic, rgba, stride) && WebPEncode(&config, &pic);
  picture.WebPPictureFree(&pic);
  if (!ok) {
    WebPMemoryWriterClear(&wrt);
    *output = nil;
    return 0;
  }
  *output = wrt.mem;
  return wrt.size;
}

// #define ENCODE_FUNC(NAME, IMPORTER)                              \
//   uint64 NAME(/* const */ in *uint8, w int, h int, bps int, float64 q, \
//               out *uint8 ) {                                   \
//     return Encode(in, w, h, bps, IMPORTER, q, 0, out);           \
//   }

func WebPEncodeRGB(/* const */ in *uint8, w int, h int, bps int, q float64 , out *uint8) uint64 {
	return Encode(in, w, h, bps, picture.WebPPictureImportRGB, q, 0, out);
}
func WebPEncodeRGBA(/* const */ in *uint8, w int, h int, bps int, q float64 , out *uint8) uint64 {
	return Encode(in, w, h, bps, picture.WebPPictureImportRGBA, q, 0, out);
}

// #define LOSSLESS_ENCODE_FUNC(NAME, IMPORTER)                                  \
//   uint64 NAME(/* const */ in *uint8, w int, h int, bps int, out *uint8 ) {      \
//     return Encode(in, w, h, bps, IMPORTER, LOSSLESS_DEFAULT_QUALITY, 1, out); \
//   }

func WebPEncodeLosslessRGB(/* const */ in *uint8, w int, h int, bps int, out *uint8 ) uint64 {
	return Encode(in, w, h, bps, picture.WebPPictureImportRGB, LOSSLESS_DEFAULT_QUALITY, 1, out);
}
func WebPEncodeLosslessRGBA(/* const */ in *uint8, w int, h int, bps int, out *uint8 ) uint64 {
	return Encode(in, w, h, bps, picture.WebPPictureImportRGBA, LOSSLESS_DEFAULT_QUALITY, 1, out);
}