// Copyright 2011 Google Inc. All Rights Reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the COPYING file in the root of the source
// tree. An additional intellectual property rights grant can be found
// in the file PATENTS. All contributing project authors may
// be found in the AUTHORS file in the root of the source tree.
package enc


// Return the encoder's version number, packed in hexadecimal using 8bits for
// each of major/minor/revision. E.g: v2.5.7 is 0x020507.
func WebPGetEncoderVersion() int {
  return (ENC_MAJ_VERSION << 16) | (ENC_MIN_VERSION << 8) | ENC_REV_VERSION;
}

func ResetSegmentHeader(/* const */ enc *VP8Encoder) {
  var hdr *VP8EncSegmentHeader = &enc.segment_hdr;
  hdr.num_segments = enc.config.Segments;
  hdr.update_map = (hdr.num_segments > 1);
  hdr.size = 0;
}

func ResetFilterHeader(/* const */ enc *VP8Encoder) {
  var hdr *VP8EncFilterHeader = &enc.filter_hdr;
  hdr.simple = 1;
  hdr.level = 0;
  hdr.sharpness = 0;
  hdr.i4x4_lf_delta = 0;
}

func ResetBoundaryPredictions(/* const */ enc *VP8Encoder) {
  // init boundary values once for all
  // Note: actually, initializing the 'preds[]' is only needed for intra4.
  var i int
  var top *uint8 = enc.preds - enc.preds_w;
  var left *uint8 = enc.preds - 1;
  for i = -1; i < 4 * enc.mb_w; i++ {
    top[i] = B_DC_PRED;
  }
  for i = 0; i < 4 * enc.mb_h; i++ {
    left[i * enc.preds_w] = B_DC_PRED;
  }
  enc.nz[-1] = 0;  // constant
}

// Mapping from config.Method to coding tools used.
//-------------------+---+---+---+---+---+---+---+
//   Method          | 0 | 1 | 2 | 3 |(4)| 5 | 6 |
//-------------------+---+---+---+---+---+---+---+
// fast probe        | x |   |   | x |   |   |   |
//-------------------+---+---+---+---+---+---+---+
// dynamic proba     | ~ | x | x | x | x | x | x |
//-------------------+---+---+---+---+---+---+---+
// fast mode analysis|[x]|[x]|   |   | x | x | x |
//-------------------+---+---+---+---+---+---+---+
// basic rd-opt      |   |   |   | x | x | x | x |
//-------------------+---+---+---+---+---+---+---+
// disto-refine i4/16| x | x | x |   |   |   |   |
//-------------------+---+---+---+---+---+---+---+
// disto-refine uv   |   | x | x |   |   |   |   |
//-------------------+---+---+---+---+---+---+---+
// rd-opt i4/16      |   |   | ~ | x | x | x | x |
//-------------------+---+---+---+---+---+---+---+
// token buffer (opt)|   |   |   | x | x | x | x |
//-------------------+---+---+---+---+---+---+---+
// Trellis           |   |   |   |   |   | x |Ful|
//-------------------+---+---+---+---+---+---+---+
// full-SNS          |   |   |   |   | x | x | x |
//-------------------+---+---+---+---+---+---+---+
func MapConfigToTools(/* const */ enc *VP8Encoder) {
  var config *config.Config = enc.config
  method := config.Method
  limit := 100 - config.PartitionLimit
  enc.method = method

  switch method {
  case 6:
    enc.rd_opt_level = RD_OPT_TRELLIS_ALL
  case 5:
    enc.rd_opt_level = RD_OPT_TRELLIS
  case 3, 4:
    enc.rd_opt_level = RD_OPT_BASIC
  default:
    enc.rd_opt_level = RD_OPT_NONE
  }

  enc.max_i4_header_bits =
      256 * 16 * 16 *                 // upper bound: up to 16bit per 4x4 block
      (limit * limit) / (100 * 100);  // ... modulated with a quadratic curve.

  // partition0 = 512k max.
  enc.mb_header_limit = score_t(256 * 510 * 8 * 1024 / (enc.mb_w * enc.mb_h))

  enc.thread_level = config.ThreadLevel;

  enc.do_search = (config.TargetSize > 0 || config.TargetPSNR > 0);
  if (!config.LowMemory) {
	enc.use_tokens = (enc.rd_opt_level >= RD_OPT_BASIC);  // need rd stats
    if (enc.use_tokens) {
      enc.num_parts = 1;  // doesn't work with multi-partition
    }
  }
}

// Memory scaling with dimensions:
//  memory (bytes) ~= 2.25 * w + 0.0625 * w * h
//
// Typical memory footprint (614x440 picture)
//              encoder: 22111
//                 info: 4368
//                preds: 17741
//          top samples: 1263
//             non-zero: 175
//             lf-stats: 0
//                total: 45658
// Transient object sizes:
//       VP8EncIterator: 3360
//         VP8ModeScore: 872
//       VP8SegmentInfo: 732
//          VP8EncProba: 18352
//              LFStats: 2048
// Picture size (yuv): 419328
func InitVP8Encoder(/* const */ config *config.Config, /*const*/ picture *picture.Picture) *VP8Encoder {
  var enc *VP8Encoder
  use_filter := (config.FilterStrength > 0) || (config.Autofilter > 0);
  mb_w := (picture.Width + 15) >> 4;
  mb_h := (picture.Height + 15) >> 4;
  preds_w := 4 * mb_w + 1;
  preds_h := 4 * mb_h + 1;
  preds_size := preds_w * preds_h * sizeof(*enc.preds);
  top_stride := mb_w * 16;
  nz_size := (mb_w + 1) * sizeof(*enc.nz) + WEBP_ALIGN_CST;
  info_size := mb_w * mb_h * sizeof(*enc.mb_info);
  samples_size :=
      2 * top_stride * sizeof(*enc.y_top)  // top-luma/u/v
      + WEBP_ALIGN_CST;                     // align all
  lf_stats_size := tenary.If(config.Autofilter, sizeof(*enc.lf_stats) + WEBP_ALIGN_CST,  0)
  top_derr_size := tenary.If(config.Quality <= ERROR_DIFFUSION_QUALITY || config.pass > 1, mb_w * sizeof(*enc.top_derr), 0)
  var mem *uint8;

  var size uint64  = uint64(sizeof(*enc)  // main struct
                        + WEBP_ALIGN_CST        // cache alignment
                        + info_size             // modes info
                        + preds_size            // prediction modes
                        + samples_size          // top/left samples
                        + top_derr_size         // top diffusion error
                        + nz_size               // coeff context bits
                        + lf_stats_size)        // autofilter stats

//   mem = (*uint8)WebPSafeMalloc(size, sizeof(*mem));
//   if (mem == nil) {
//     WebPEncodingSetError(picture, VP8_ENC_ERROR_OUT_OF_MEMORY);
//     return nil;
//   }
	mem = make([]uint8, size)

	enc = (*VP8Encoder)(mem)
	mem = (*uint8)WEBP_ALIGN(mem + sizeof(*enc));
	stdlib.Memset(enc, 0, sizeof(*enc));
	enc.num_parts = 1 << config.Partitions;
	enc.mb_w = mb_w;
	enc.mb_h = mb_h;
	enc.preds_w = preds_w;
	enc.mb_info = (*VP8MBInfo)mem;
	mem += info_size;
	enc.preds = mem + 1 + enc.preds_w;
	mem += preds_size;
	enc.nz = 1 + (*uint32)WEBP_ALIGN(mem);
	mem += nz_size;
	enc.lf_stats = lf_stats_size ? (*LFStats)WEBP_ALIGN(mem) : nil;
	mem += lf_stats_size;

  // top samples (all 16-aligned)
  mem = (*uint8)WEBP_ALIGN(mem);
  enc.y_top = mem;
  enc.uv_top = enc.y_top + top_stride;
  mem += 2 * top_stride;
  enc.top_derr = top_derr_size ? (*DError)mem : nil;
  mem += top_derr_size;
  assert.Assert(mem <= (*uint8)enc + size);

  enc.config = config;
  enc.profile = use_filter ? ((config.FilterType == 1) ? 0 : 1) : 2;
  enc.pic = picture;
  enc.percent = 0;

  MapConfigToTools(enc);
  VP8EncDspInit();
  VP8DefaultProbas(enc);
  ResetSegmentHeader(enc);
  ResetFilterHeader(enc);
  ResetBoundaryPredictions(enc);
  VP8EncDspCostInit();
  VP8EncInitAlpha(enc);

  // lower quality means smaller output . we modulate a little the page
  // size based on quality. This is just a crude 1rst-order prediction.
  {
    var float64 scale = 1.0 + config.Quality * 5.0 / 100.0;  // in [1,6]
    VP8TBufferInit(&enc.tokens, (int)(mb_w * mb_h * 4 * scale));
  }
  return enc;
}

func DeleteVP8Encoder(enc *VP8Encoder) int {
  ok := 1;
  if (enc != nil) {
    ok = VP8EncDeleteAlpha(enc);
    VP8TBufferClear(&enc.tokens);
  }
  return ok;
}

func StoreStats(/* const */ enc *VP8Encoder) {
	WebPReportProgress(enc.pic, 100, &enc.percent);  // done!
}



func WebPEncode(/* const */ config *config.Config, pic *picture.Picture) int {
  ok := 0;
  if pic == nil { return 0  }

  pic.ErrorCode = VP8_ENC_OK;  // all ok so far
  if (config == nil) {          // bad params
    return WebPEncodingSetError(pic, VP8_ENC_ERROR_nil_PARAMETER);
  }
  err := config.Validate()
  if (err != nil) { //TODO: just return err
    return WebPEncodingSetError(pic, VP8_ENC_ERROR_INVALID_CONFIGURATION);
  }
  if !WebPValidatePicture(pic) { return 0  }
  if (pic.width > WEBP_MAX_DIMENSION || pic.height > WEBP_MAX_DIMENSION) {
    return WebPEncodingSetError(pic, VP8_ENC_ERROR_BAD_DIMENSION);
  }

  if pic.stats != nil { stdlib.Memset(pic.stats, 0, sizeof(*pic.stats)) }

  if (!config.Lossless) {
    enc *VP8Encoder = nil;

    if (pic.use_argb || pic.y == nil || pic.u == nil || pic.v == nil) {
      // Make sure we have YUVA samples.
      if (config.UseSharpYUV || (config.Preprocessing & 4)) {
        if (!picture.WebPPictureSharpARGBToYUVA(pic)) {
          return 0;
        }
      } else {
        float64 dithering = 0.0;
        if (config.Preprocessing & 2) {
          const float64 x = config.Quality / 100.0;
          const float64 x2 = x * x;
          // slowly decreasing from max dithering at low quality (q.0)
          // to 0.5 dithering amplitude at high quality (q.100)
          dithering = 1.0 + (0.5 - 1.0) * x2 * x2;
        }
        if (!picture.WebPPictureARGBToYUVADithered(pic, colorspace.WEBP_YUV420, dithering)) {
          return 0;
        }
      }
    }

    if (!config.Exact) {
      WebPCleanupTransparentArea(pic);
    }

    enc = InitVP8Encoder(config, pic);
    if enc == nil {
		return 0  // pic.error is already set.
	}
    // Note: each of the tasks below account for 20% in the progress report.
    ok = VP8EncAnalyze(enc);

    // Analysis is done, proceed to actual coding.
    ok = ok && VP8EncStartAlpha(enc);  // possibly done in parallel
    if (!enc.use_tokens) {
      ok = ok && VP8EncLoop(enc);
    } else {
      ok = ok && VP8EncTokenLoop(enc);
    }
    ok = ok && VP8EncFinishAlpha(enc);

    ok = ok && VP8EncWrite(enc);
    StoreStats(enc);
    ok &= DeleteVP8Encoder(enc);  // must always be called, even if !ok
  } else {
    // Make sure we have ARGB samples.
    if (pic.argb == nil && !picture.WebPPictureYUVAToARGB(pic)) {
      return 0;
    }

    if (!config.Exact) {
      WebPReplaceTransparentPixels(pic, 0x000000);
    }

    ok = VP8LEncodeImage(config, pic);  // Sets pic.error in case of problem.
  }

  return ok;
}
