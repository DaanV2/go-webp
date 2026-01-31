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
  WEBP_MUX_DISPOSE_NONE WebPMuxAnimDispose = iota       // Do not dispose.
  WEBP_MUX_DISPOSE_BACKGROUND  // Dispose to background color.
)

// Blend operation (animation only). Indicates how transparent pixels of the
// current frame are blended with those of the previous canvas.
type WebPMuxAnimBlend int

const (
	WEBP_MUX_BLEND WebPMuxAnimBlend = iota    // Blend.
  	WEBP_MUX_NO_BLEND  // Do not blend.
)