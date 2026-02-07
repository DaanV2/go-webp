package config

import "errors"

// Compression parameters.
type WebPConfig struct {
	lossless int // Lossless encoding (0=lossy(default), 1=lossless).
	// between 0 and 100. For lossy, 0 gives the smallest
	// size and 100 the largest. For lossless, this
	// parameter is the amount of effort put into the
	// compression: 0 is the fastest but gives larger
	// files compared to the slowest, but best, 100.
	quality float64
	method  int // quality/speed trade-off (0=fast, 6=slower-better)

	image_hint WebPImageHint // Hint for image type (lossless only for now).

	// if non-zero, set the desired target size in bytes.
	// Takes precedence over the 'compression' parameter.
	target_size int
	// if non-zero, specifies the minimal distortion to
	// try to achieve. Takes precedence over target_size.
	target_PSNR      float64
	segments         int // maximum number of segments to use, in [1..4]
	sns_strength     int // Spatial Noise Shaping. 0=off, 100=maximum.
	filter_strength  int // range: [0 = off .. 100 = strongest]
	filter_sharpness int // range: [0 = off .. 7 = least sharp]
	// filtering type: 0 = simple, 1 = strong (only used
	// if filter_strength > 0 or autofilter > 0)
	filter_type       int
	autofilter        int // Auto adjust filter's strength [0 = off, 1 = on]
	alpha_compression int // Algorithm for encoding the alpha plane (0 = none, // 1 = compressed with WebP lossless). Default is 1.
	// Predictive filtering method for alpha plane.
	//  0: none, 1: fast, 2: best. Default if 1.
	alpha_filtering int
	// Between 0 (smallest size) and 100 (lossless).
	// Default is 100.
	alpha_quality int
	pass          int // number of entropy-analysis passes (in [1..10]).

	// if true, export the compressed picture back.
	// In-loop filtering is not applied.
	show_compressed int
	// preprocessing filter:
	// 0=none, 1=segment-smooth, 2=pseudo-random dithering
	preprocessing int
	// log2(number of token partitions) in [0..3]. Default
	// is set to 0 for easier progressive decoding.
	partitions int
	// quality degradation allowed to fit the 512k limit
	// on prediction modes coding (0: no degradation, // 100: maximum possible degradation).
	partition_limit int
	// If true, compression parameters will be remapped
	// to better match the expected output size from
	// JPEG compression. Generally, the output size will
	// be similar but the degradation will be lower.
	emulate_jpeg_size int
	thread_level      int // If non-zero, try and use multi-threaded encoding.
	low_memory        int // If set, reduce memory usage (but increase CPU use).

	// Near lossless encoding [0 = max loss .. 100 = off
	// (default)].
	near_lossless int

	// if non-zero, preserve the exact RGB values under
	// transparent area. Otherwise, discard this invisible
	// RGB information for better compression. The default
	// value is 0.
	exact int

	use_delta_palette int // reserved
	use_sharp_yuv     int // if needed, use sharp (and slow) RGB.YUV conversion

	qmin int // minimum permissible quality factor
	qmax int // maximum permissible quality factor
}

// Should always be called, to initialize a fresh WebPConfig structure before
// modification. Returns false in case of version mismatch. WebPConfigInit()
// must have succeeded before using the 'config' object.
// Note that the default values are lossless=0 and quality=75.
func WebPConfigInit(config *WebPConfig) error {
	return WebPConfigInitInternal(config, WEBP_PRESET_DEFAULT, 75.0, WEBP_ENCODER_ABI_VERSION)
}

// This function will initialize the configuration according to a predefined
// set of parameters (referred to by 'preset') and a given quality factor.
// This function can be called as a replacement to WebPConfigInit(). Will
// return false in case of error.
func WebPConfigPreset(config *WebPConfig, preset WebPPreset, quality float64) error {
	return WebPConfigInitInternal(config, preset, quality, WEBP_ENCODER_ABI_VERSION)
}

// Internal, version-checked, entry point
func WebPConfigInitInternal(config *WebPConfig, preset WebPPreset, quality float64, version int) error {
	if WEBP_ABI_IS_INCOMPATIBLE(version, WEBP_ENCODER_ABI_VERSION) {
		return nil // caller/system version mismatch!
	}
	if config == nil {
		return nil
	}

	config.quality = quality
	config.target_size = 0
	config.target_PSNR = 0.0
	config.method = 4
	config.sns_strength = 50
	config.filter_strength = 60 // mid-filtering
	config.filter_sharpness = 0
	config.filter_type = 1 // default: strong (so U/V is filtered too)
	config.partitions = 0
	config.segments = 4
	config.pass = 1
	config.qmin = 0
	config.qmax = 100
	config.show_compressed = 0
	config.preprocessing = 0
	config.autofilter = 0
	config.partition_limit = 0
	config.alpha_compression = 1
	config.alpha_filtering = 1
	config.alpha_quality = 100
	config.lossless = 0
	config.exact = 0
	config.image_hint = WEBP_HINT_DEFAULT
	config.emulate_jpeg_size = 0
	config.thread_level = 0
	config.low_memory = 0
	config.near_lossless = 100
	config.use_sharp_yuv = 0

	// TODO(skal): tune.
	switch preset {
	case WEBP_PRESET_PICTURE:
		config.sns_strength = 80
		config.filter_sharpness = 4
		config.filter_strength = 35
		config.preprocessing &= ^2 // no dithering
		break
	case WEBP_PRESET_PHOTO:
		config.sns_strength = 80
		config.filter_sharpness = 3
		config.filter_strength = 30
		config.preprocessing |= 2
		break
	case WEBP_PRESET_DRAWING:
		config.sns_strength = 25
		config.filter_sharpness = 6
		config.filter_strength = 10
		break
	case WEBP_PRESET_ICON:
		config.sns_strength = 0
		config.filter_strength = 0 // disable filtering to retain sharpness
		config.preprocessing &= ^2 // no dithering
		break
	case WEBP_PRESET_TEXT:
		config.sns_strength = 0
		config.filter_strength = 0 // disable filtering to retain sharpness
		config.preprocessing &= ^2 // no dithering
		config.segments = 2
		break
	case WEBP_PRESET_DEFAULT:
	default:
		break
	}
	return config.Validate()
}

// Returns true if 'config' is non-nil and all configuration parameters are
// within their valid ranges.
func (config *WebPConfig) Validate() error {
	if config == nil {
		return errors.New("config is nil")
	}
	if config.quality < 0 || config.quality > 100 {
		return errors.New("quality must be between 0 and 100")
	}
	if config.target_size < 0 {
		return errors.New("target_size must be non-negative")
	}
	if config.target_PSNR < 0 {
		return errors.New("target_PSNR must be non-negative")
	}
	if config.method < 0 || config.method > 6 {
		return errors.New("method must be between 0 and 6")
	}
	if config.segments < 1 || config.segments > 4 {
		return errors.New("segments must be between 1 and 4")
	}
	if config.sns_strength < 0 || config.sns_strength > 100 {
		return errors.New("sns_strength must be between 0 and 100")
	}
	if config.filter_strength < 0 || config.filter_strength > 100 {
		return errors.New("filter_strength must be between 0 and 100")
	}
	if config.filter_sharpness < 0 || config.filter_sharpness > 7 {
		return errors.New("filter_sharpness must be between 0 and 7")
	}
	if config.filter_type < 0 || config.filter_type > 1 {
		return errors.New("filter_type must be 0 or 1")
	}
	if config.autofilter < 0 || config.autofilter > 1 {
		return errors.New("autofilter must be 0 or 1")
	}
	if config.pass < 1 || config.pass > 10 {
		return errors.New("pass must be between 1 and 10")
	}
	if config.qmin < 0 || config.qmax > 100 || config.qmin > config.qmax {
		return errors.New("qmin/qmax must be in [0,100] and qmin <= qmax")
	}
	if config.show_compressed < 0 || config.show_compressed > 1 {
		return errors.New("show_compressed must be 0 or 1")
	}
	if config.preprocessing < 0 || config.preprocessing > 7 {
		return errors.New("preprocessing must be between 0 and 7")
	}
	if config.partitions < 0 || config.partitions > 3 {
		return errors.New("partitions must be between 0 and 3")
	}
	if config.partition_limit < 0 || config.partition_limit > 100 {
		return errors.New("partition_limit must be between 0 and 100")
	}
	if config.alpha_compression < 0 {
		return errors.New("alpha_compression must be non-negative")
	}
	if config.alpha_filtering < 0 {
		return errors.New("alpha_filtering must be non-negative")
	}
	if config.alpha_quality < 0 || config.alpha_quality > 100 {
		return errors.New("alpha_quality must be between 0 and 100")
	}
	if config.lossless < 0 || config.lossless > 1 {
		return errors.New("lossless must be 0 or 1")
	}
	if config.near_lossless < 0 || config.near_lossless > 100 {
		return errors.New("near_lossless must be between 0 and 100")
	}
	if config.image_hint >= WEBP_HINT_LAST {
		return errors.New("image_hint must be less than WEBP_HINT_LAST")
	}
	if config.emulate_jpeg_size < 0 || config.emulate_jpeg_size > 1 {
		return errors.New("emulate_jpeg_size must be 0 or 1")
	}
	if config.thread_level < 0 || config.thread_level > 1 {
		return errors.New("thread_level must be 0 or 1")
	}
	if config.low_memory < 0 || config.low_memory > 1 {
		return errors.New("low_memory must be 0 or 1")
	}
	if config.exact < 0 || config.exact > 1 {
		return errors.New("exact must be 0 or 1")
	}
	if config.use_sharp_yuv < 0 || config.use_sharp_yuv > 1 {
		return errors.New("use_sharp_yuv must be 0 or 1")
	}

	return nil
}
