package enc

import (
	"github.com/daanv2/go-webp/pkg/color/colorspace"
	"github.com/daanv2/go-webp/pkg/config"
	"github.com/daanv2/go-webp/pkg/picture"
	"github.com/daanv2/go-webp/pkg/stdlib"
)

// Main encoding call, after config and picture have been initialized.
// 'picture' must be less than 16384x16384 in dimension (cf WEBP_MAX_DIMENSION),
// and the 'config' object must be a valid one.
// Returns false in case of error, true otherwise.
// In case of error, picture.ErrorCode is updated accordingly.
// 'picture' can hold the source samples in both YUV(A) or ARGB input, depending
// on the value of 'picture.UseARGB'. It is highly recommended to use
// the former for lossy encoding, and the latter for lossless encoding
// (when config.Lossless is true). Automatic conversion from one format to
// another is provided but they both incur some loss.
func WebPEncode( /* const */ config *config.Config, pic *picture.Picture) int {
	ok := 0
	if pic == nil {
		return 0
	}

	pic.ErrorCode = picture.ENC_OK // all ok so far
	if config == nil {         // bad params
		return picture.WebPEncodingSetError(pic, ENC_ERROR_nil_PARAMETER)
	}
	err := config.Validate()
	if err != nil { //TODO: just return err
		return WebPEncodingSetError(pic, ENC_ERROR_INVALID_CONFIGURATION)
	}
	if !WebPValidatePicture(pic) {
		return 0
	}
	if pic.width > WEBP_MAX_DIMENSION || pic.height > WEBP_MAX_DIMENSION {
		return WebPEncodingSetError(pic, ENC_ERROR_BAD_DIMENSION)
	}

	if pic.stats != nil {
		stdlib.Memset(pic.stats, 0, sizeof(*pic.stats))
	}

	if !config.Lossless {
		enc * VP8Encoder = nil

		if pic.use_argb || pic.y == nil || pic.u == nil || pic.v == nil {
			// Make sure we have YUVA samples.
			if config.UseSharpYUV || (config.Preprocessing & 4) {
				if !picture.WebPPictureSharpARGBToYUVA(pic) {
					return 0
				}
			} else {
				var dithering float64 = 0.0
				if config.Preprocessing & 2 {
					var float64 x = config.Quality / 100.0
					var float64 x2 = x * x
					// slowly decreasing from max dithering at low quality (q.0)
					// to 0.5 dithering amplitude at high quality (q.100)
					dithering = 1.0 + (0.5-1.0)*x2*x2
				}
				if !picture.WebPPictureARGBToYUVADithered(pic, colorspace.WEBP_YUV420, dithering) {
					return 0
				}
			}
		}

		if !config.Exact {
			WebPCleanupTransparentArea(pic)
		}

		enc = InitVP8Encoder(config, pic)
		if enc == nil {
			return 0 // pic.error is already set.
		}
		// Note: each of the tasks below account for 20% in the progress report.
		ok = VP8EncAnalyze(enc)

		// Analysis is done, proceed to actual coding.
		ok = ok && VP8EncStartAlpha(enc) // possibly done in parallel
		if !enc.use_tokens {
			ok = ok && VP8EncLoop(enc)
		} else {
			ok = ok && VP8EncTokenLoop(enc)
		}
		ok = ok && VP8EncFinishAlpha(enc)

		ok = ok && VP8EncWrite(enc)
		StoreStats(enc)
		ok &= DeleteVP8Encoder(enc) // must always be called, even if !ok
	} else {
		// Make sure we have ARGB samples.
		if pic.argb == nil && !picture.WebPPictureYUVAToARGB(pic) {
			return 0
		}

		if !config.Exact {
			WebPReplaceTransparentPixels(pic, 0x000000)
		}

		ok = VP8LEncodeImage(config, pic) // Sets pic.error in case of problem.
	}

	return ok
}
