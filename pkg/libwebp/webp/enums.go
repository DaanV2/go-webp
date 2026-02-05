package webp

// VP8X Feature Flags.
type WebPFeatureFlags int

const (
	ANIMATION_FLAG WebPFeatureFlags = 0x00000002
	XMP_FLAG       WebPFeatureFlags = 0x00000004
	EXIF_FLAG      WebPFeatureFlags = 0x00000008
	ALPHA_FLAG     WebPFeatureFlags = 0x00000010
	ICCP_FLAG      WebPFeatureFlags = 0x00000020

	ALL_VALID_FLAGS WebPFeatureFlags = 0x0000003e
)

// Dispose method (animation only). Indicates how the area used by the current
// frame is to be treated before rendering the next frame on the canvas.
type WebPMuxAnimDispose int

const (
	WEBP_MUX_DISPOSE_NONE       WebPMuxAnimDispose = iota // Do not dispose.
	WEBP_MUX_DISPOSE_BACKGROUND                           // Dispose to background color.
)

// Blend operation (animation only). Indicates how transparent pixels of the
// current frame are blended with those of the previous canvas.
type WebPMuxAnimBlend int

const (
	WEBP_MUX_BLEND    WebPMuxAnimBlend = iota // Blend.
	WEBP_MUX_NO_BLEND                         // Do not blend.
)

// Colorspaces
// Note: the naming describes the byte-ordering of packed samples in memory.
// For instance, MODE_BGRA relates to samples ordered as B,G,R,A,B,G,R,A,...
// Non-capital names (e.g.:MODE_Argb) relates to pre-multiplied RGB channels.
// RGBA-4444 and RGB-565 colorspaces are represented by following byte-order:
// RGBA-4444: [r3 r2 r1 r0 g3 g2 g1 g0], [b3 b2 b1 b0 a3 a2 a1 a0], ...
// RGB-565: [r4 r3 r2 r1 r0 g5 g4 g3], [g2 g1 g0 b4 b3 b2 b1 b0], ...
// In the case WEBP_SWAP_16BITS_CSP is defined, the bytes are swapped for
// these two modes:
// RGBA-4444: [b3 b2 b1 b0 a3 a2 a1 a0], [r3 r2 r1 r0 g3 g2 g1 g0], ...
// RGB-565: [g2 g1 g0 b4 b3 b2 b1 b0], [r4 r3 r2 r1 r0 g5 g4 g3], ...
type WEBP_CSP_MODE int

const (
	MODE_RGB       WEBP_CSP_MODE = 0
	MODE_RGBA      WEBP_CSP_MODE = 1
	MODE_BGR       WEBP_CSP_MODE = 2
	MODE_BGRA      WEBP_CSP_MODE = 3
	MODE_ARGB      WEBP_CSP_MODE = 4
	MODE_RGBA_4444 WEBP_CSP_MODE = 5
	MODE_RGB_565   WEBP_CSP_MODE = 6
	// RGB-premultiplied transparent modes (alpha value is preserved)
	MODE_rgbA      WEBP_CSP_MODE = 7
	MODE_bgrA      WEBP_CSP_MODE = 8
	MODE_Argb      WEBP_CSP_MODE = 9
	MODE_rgbA_4444 WEBP_CSP_MODE = 10
	// YUV modes must come after RGB ones.
	MODE_YUV  WEBP_CSP_MODE = 11
	MODE_YUVA WEBP_CSP_MODE = 12 // yuv 4:2:0
	MODE_LAST WEBP_CSP_MODE = 13
)

// Some useful macros:
func WebPIsPremultipliedMode(mode WEBP_CSP_MODE) bool {
	return (mode == MODE_rgbA || mode == MODE_bgrA || mode == MODE_Argb ||
		mode == MODE_rgbA_4444)
}

func (mode WEBP_CSP_MODE) IsPremultipliedMode() bool {
	return WebPIsPremultipliedMode(mode)
}

func WebPIsAlphaMode(mode WEBP_CSP_MODE) bool {
	return (mode == MODE_RGBA || mode == MODE_BGRA || mode == MODE_ARGB ||
		mode == MODE_RGBA_4444 || mode == MODE_YUVA ||
		WebPIsPremultipliedMode(mode))
}

func (mode WEBP_CSP_MODE) IsAlphaMode() bool {
	return WebPIsAlphaMode(mode)
}

func WebPIsRGBMode(mode WEBP_CSP_MODE) bool {
	return (mode < MODE_YUV)
}

func (mode WEBP_CSP_MODE) IsRGBMode() bool {
	return WebPIsRGBMode(mode)
}

//------------------------------------------------------------------------------
// Enumeration of the status codes

// Error codes
type WebPMuxError int

const (
	WEBP_MUX_OK               WebPMuxError = 1
	WEBP_MUX_NOT_FOUND        WebPMuxError = 0
	WEBP_MUX_INVALID_ARGUMENT WebPMuxError = -1
	WEBP_MUX_BAD_DATA         WebPMuxError = -2
	WEBP_MUX_MEMORY_ERROR     WebPMuxError = -3
	WEBP_MUX_NOT_ENOUGH_DATA  WebPMuxError = -4
)

type WebPDemuxState int

const (
	WEBP_DEMUX_PARSE_ERROR    WebPDemuxState = -1 // An error occurred while parsing.
	WEBP_DEMUX_PARSING_HEADER WebPDemuxState = 0  // Not enough data to parse full header.
	WEBP_DEMUX_PARSED_HEADER  WebPDemuxState = 1  // Header parsing complete, // data may be available.
	WEBP_DEMUX_DONE           WebPDemuxState = 2  // Entire file has been parsed.
)

type WebPFormatFeature int

const (
	// bit-wise combination of WebPFeatureFlags
	// corresponding to the 'VP8X' chunk (if present).
	WEBP_FF_FORMAT_FLAGS WebPFormatFeature = iota
	WEBP_FF_CANVAS_WIDTH
	WEBP_FF_CANVAS_HEIGHT
	// only relevant for animated file
	WEBP_FF_LOOP_COUNT
	// idem.
	WEBP_FF_BACKGROUND_COLOR
	// Number of frames present in the demux object.
	// In case of a partial demux, this is the number
	// of frames seen so far, with the last frame
	// possibly being partial.
	WEBP_FF_FRAME_COUNT
)

// Image characteristics hint for the underlying encoder.
type WebPImageHint int

const (
	WEBP_HINT_DEFAULT WebPImageHint = iota // default preset.
	WEBP_HINT_PICTURE                      // digital picture, like portrait, inner shot
	WEBP_HINT_PHOTO                        // outdoor photograph, with natural lighting
	WEBP_HINT_GRAPH                        // Discrete tone image (graph, map-tile etc).
	WEBP_HINT_LAST
)

// Enumerate some predefined settings for WebPConfig, depending on the type
// of source picture. These presets are used when calling WebPConfigPreset().
type WebPPreset int

const (
	WEBP_PRESET_DEFAULT WebPPreset = iota // default preset.
	WEBP_PRESET_PICTURE                   // digital picture, like portrait, inner shot
	WEBP_PRESET_PHOTO                     // outdoor photograph, with natural lighting
	WEBP_PRESET_DRAWING                   // hand or line drawing, with high-contrast details
	WEBP_PRESET_ICON                      // small-sized colorful images
	WEBP_PRESET_TEXT                      // text-like
)

// Color spaces.
type WebPEncCSP int

const (
	// chroma sampling
	WEBP_YUV420        WebPEncCSP = 0 // 4:2:0
	WEBP_YUV420A       WebPEncCSP = 4 // alpha channel variant
	WEBP_CSP_UV_MASK   WebPEncCSP = 3 // bit-mask to get the UV sampling factors
	WEBP_CSP_ALPHA_BIT WebPEncCSP = 4 // bit that is set if alpha is present
)

// Encoding error conditions.
type WebPEncodingError int

const (
	VP8_ENC_OK                            WebPEncodingError = iota
	VP8_ENC_ERROR_OUT_OF_MEMORY                             // memory error allocating objects
	VP8_ENC_ERROR_BITSTREAM_OUT_OF_MEMORY                   // memory error while flushing bits
	VP8_ENC_ERROR_nil_PARAMETER                             // a pointer parameter is nil
	VP8_ENC_ERROR_INVALID_CONFIGURATION                     // configuration is invalid
	VP8_ENC_ERROR_BAD_DIMENSION                             // picture has invalid width/height
	VP8_ENC_ERROR_PARTITION0_OVERFLOW                       // partition is bigger than 512k
	VP8_ENC_ERROR_PARTITION_OVERFLOW                        // partition is bigger than 16M
	VP8_ENC_ERROR_BAD_WRITE                                 // error while flushing bytes
	VP8_ENC_ERROR_FILE_TOO_BIG                              // file is bigger than 4G
	VP8_ENC_ERROR_USER_ABORT                                // abort request by user
	VP8_ENC_ERROR_LAST                                      // list terminator. always last.
)
