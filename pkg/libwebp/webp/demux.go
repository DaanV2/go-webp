// Copyright 2012 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.

package webp

// Code Example: Demuxing WebP data to extract all the frames, ICC profile
// and EXIF/XMP metadata.
/*
  demux *WebPDemuxer = WebPDemux(&webp_data)

  width := WebPDemuxGetI(demux, WEBP_FF_CANVAS_WIDTH)
  height := WebPDemuxGetI(demux, WEBP_FF_CANVAS_HEIGHT)
  // ... (Get information about the features present in the WebP file).
  flags := WebPDemuxGetI(demux, WEBP_FF_FORMAT_FLAGS)

  // ... (Iterate over all frames).
   var iter WebPIterator
  if (WebPDemuxGetFrame(demux, 1, &iter)) {
    for {
      // ... (Consume 'iter'; e.g. Decode 'iter.fragment' with WebPDecode(), // ... and get other frame properties like width, height, offsets etc.
      // ... see 'struct WebPIterator' below for more info).
    } while (WebPDemuxNextFrame(&iter))
    WebPDemuxReleaseIterator(&iter)
  }

  // ... (Extract metadata).
   var chunk_iter WebPChunkIterator
  if flags & ICCP_FLAG { WebPDemuxGetChunk(demux, "ICCP", 1, &chunk_iter) }
  // ... (Consume the ICC profile in 'chunk_iter.chunk').
  WebPDemuxReleaseChunkIterator(&chunk_iter)
  if flags & EXIF_FLAG { WebPDemuxGetChunk(demux, "EXIF", 1, &chunk_iter) }
  // ... (Consume the EXIF metadata in 'chunk_iter.chunk').
  WebPDemuxReleaseChunkIterator(&chunk_iter)
  if flags & XMP_FLAG { WebPDemuxGetChunk(demux, "XMP ", 1, &chunk_iter) }
  // ... (Consume the XMP metadata in 'chunk_iter.chunk').
  WebPDemuxReleaseChunkIterator(&chunk_iter)
  WebPDemuxDelete(demux)
*/

// Parses the full WebP file given by 'data'. For single images the WebP file
// header alone or the file header and the chunk header may be absent.
// Returns a WebPDemuxer object on successful parse, nil otherwise.
func WebPDemuxer( /* const */ data *WebPData) *WebPDemux {
	return WebPDemuxInternal(data, 0, nil, WEBP_DEMUX_ABI_VERSION)
}

// Parses the possibly incomplete WebP file given by 'data'.
// If 'state' is non-nil it will be set to indicate the status of the demuxer.
// Returns nil in case of error or if there isn't enough data to start parsing
// and a WebPDemuxer object on successful parse.
// Note that WebPDemuxer keeps internal pointers to 'data' memory segment.
// If this data is volatile, the demuxer object should be deleted (by calling
// WebPDemuxDelete()) and WebPDemuxPartial() called again on the new data.
// This is usually an inexpensive operation.
func WebPDemuxPartial( /* const */ data *WebPData, state *WebPDemuxState) *WebPDemuxer {
	return WebPDemuxInternal(data, 1, state, WEBP_DEMUX_ABI_VERSION)
}

//------------------------------------------------------------------------------
// WebPAnimDecoder API
//
// This API allows decoding (possibly) animated WebP images.
//
// Code Example:
/*
   var dec_options WebPAnimDecoderOptions
  WebPAnimDecoderOptionsInit(&dec_options)
  // Tune 'dec_options' as needed.
  dec *WebPAnimDecoder = WebPAnimDecoderNew(webp_data, &dec_options)
   var anim_info WebPAnimInfo
  WebPAnimDecoderGetInfo(dec, &anim_info)
  for i := 0; i < anim_info.loop_count; i++ {
    while (WebPAnimDecoderHasMoreFrames(dec)) {
      buf *uint8
      var timestamp int
      WebPAnimDecoderGetNext(dec, &buf, &timestamp)
      // ... (Render 'buf' based on 'timestamp').
      // ... (Do NOT free 'buf', as it is owned by 'dec').
    }
    WebPAnimDecoderReset(dec)
  }
  var demuxer *WebPDemuxer = WebPAnimDecoderGetDemuxer(dec)
  // ... (Do something using 'demuxer'; e.g. get EXIF/XMP/ICC data).
  WebPAnimDecoderDelete(dec)
*/

// Global options.
type WebPAnimDecoderOptions struct {
	// Output colorspace. Only the following modes are supported:
	// MODE_RGBA, MODE_BGRA, MODE_rgbA and MODE_bgrA.
	color_mode  WEBP_CSP_MODE
	use_threads int       // If true, use multi-threaded decoding.
	padding     [7]uint32 // Padding for later use.
}

// Should always be called, to initialize a fresh WebPAnimDecoderOptions
// structure before modification. Returns false in case of version mismatch.
// WebPAnimDecoderOptionsInit() must have succeeded before using the
// 'dec_options' object.
func WebPAnimDecoderOptionsInit(dec_options *WebPAnimDecoderOptions) int {
	return WebPAnimDecoderOptionsInitInternal(dec_options, WEBP_DEMUX_ABI_VERSION)
}

// Creates and initializes a WebPAnimDecoder object.
// Parameters:
//   webp_data - (in) WebP bitstream. This should remain unchanged during the
//                    lifetime of the output WebPAnimDecoder object.
//   dec_options - (in) decoding options. Can be passed nil to choose
//                      reasonable defaults (in particular, color mode MODE_RGBA
//                      will be picked).
// Returns:
//   A pointer to the newly created WebPAnimDecoder object, or nil in case of
//   parsing error, invalid option or memory error.
func WebPAnimDecoderNew( /* const */ webp_data *WebPData /*const*/, dec_options *WebPAnimDecoderOptions) *WebPAnimDecoder {
	return WebPAnimDecoderNewInternal(webp_data, dec_options, WEBP_DEMUX_ABI_VERSION)
}

// Global information about the animation..
type WebPAnimInfo struct {
	canvas_width  uint32
	canvas_height uint32
	loop_count    uint32
	bgcolor       uint32
	frame_count   uint32
	pad           [4]uint32 // padding for later use
}
