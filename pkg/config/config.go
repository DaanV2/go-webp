package config

import (
	"errors"
)

// Compression parameters.
type Config struct {
	Lossless int // Lossless encoding (0=lossy(default), 1=lossless).
	// between 0 and 100. For lossy, 0 gives the smallest
	// size and 100 the largest. For lossless, this
	// parameter is the amount of effort put into the
	// compression: 0 is the fastest but gives larger
	// files compared to the slowest, but best, 100.
	Quality float64
	Method  int // quality/speed trade-off (0=fast, 6=slower-better)

	ImageHint ImageHint // Hint for image type (lossless only for now).

	// if non-zero, set the desired target size in bytes.
	// Takes precedence over the 'compression' parameter.
	TargetSize int
	// if non-zero, specifies the minimal distortion to
	// try to achieve. Takes precedence over target_size.
	TargetPSNR      float64
	Segments        int // maximum number of segments to use, in [1..4]
	SnsStrength     int // Spatial Noise Shaping. 0=off, 100=maximum.
	FilterStrength  int // range: [0 = off .. 100 = strongest]
	FilterSharpness int // range: [0 = off .. 7 = least sharp]
	// filtering type: 0 = simple, 1 = strong (only used
	// if filter_strength > 0 or autofilter > 0)
	FilterType       int
	Autofilter       int // Auto adjust filter's strength [0 = off, 1 = on]
	AlphaCompression int // Algorithm for encoding the alpha plane (0 = none, // 1 = compressed with WebP lossless). Default is 1.
	// Predictive filtering method for alpha plane.
	//  0: none, 1: fast, 2: best. Default if 1.
	AlphaFiltering int
	// Between 0 (smallest size) and 100 (lossless).
	// Default is 100.
	AlphaQuality int
	Pass         int // number of entropy-analysis passes (in [1..10]).

	// if true, export the compressed picture back.
	// In-loop filtering is not applied.
	ShowCompressed int
	// preprocessing filter:
	// 0=none, 1=segment-smooth, 2=pseudo-random dithering
	Preprocessing int
	// log2(number of token partitions) in [0..3]. Default
	// is set to 0 for easier progressive decoding.
	Partitions int
	// quality degradation allowed to fit the 512k limit
	// on prediction modes coding (0: no degradation, // 100: maximum possible degradation).
	PartitionLimit int
	// If true, compression parameters will be remapped
	// to better match the expected output size from
	// JPEG compression. Generally, the output size will
	// be similar but the degradation will be lower.
	EmulateJpegSize int
	ThreadLevel     int // If non-zero, try and use multi-threaded encoding.
	LowMemory       int // If set, reduce memory usage (but increase CPU use).

	// Near lossless encoding [0 = max loss .. 100 = off
	// (default)].
	NearLossless int

	// if non-zero, preserve the exact RGB values under
	// transparent area. Otherwise, discard this invisible
	// RGB information for better compression. The default
	// value is 0.
	Exact int

	UseDeltaPalette int // reserved
	UseSharpYUV     int // if needed, use sharp (and slow) RGB.YUV conversion

	Qmin int // minimum permissible quality factor
	Qmax int // maximum permissible quality factor
}

// Should always be called, to initialize a fresh Config structure before
// modification. Returns false in case of version mismatch. WebPConfigInit()
// must have succeeded before using the 'config' object.
// Note that the default values are lossless=0 and quality=75.
func ConfigInit(config *Config) error {
	return config.Init()
}

// Should always be called, to initialize a fresh Config structure before
// modification. Returns false in case of version mismatch. WebPConfigInit()
// must have succeeded before using the 'config' object.
// Note that the default values are lossless=0 and quality=75.
func (config *Config) Init() error {
	return config.init(WEBP_PRESET_DEFAULT, 75.0)
}

// This function will initialize the configuration according to a predefined
// set of parameters (referred to by 'preset') and a given quality factor.
// This function can be called as a replacement to WebPConfigInit(). Will
// return false in case of error.
func ConfigPreset(config *Config, preset Preset, quality float64) error {
	return config.init(preset, quality)
}

// This function will initialize the configuration according to a predefined
// set of parameters (referred to by 'preset') and a given quality factor.
// This function can be called as a replacement to WebPConfigInit(). Will
// return false in case of error.
func (config *Config) InitPreset(preset Preset, quality float64) error {
	return config.init(preset, quality)
}

// Internal, version-checked, entry point
func (config *Config) init(preset Preset, quality float64) error {
	if config == nil {
		return nil
	}

	config.Quality = quality
	config.TargetSize = 0
	config.TargetPSNR = 0.0
	config.Method = 4
	config.SnsStrength = 50
	config.FilterStrength = 60 // mid-filtering
	config.FilterSharpness = 0
	config.FilterType = 1 // default: strong (so U/V is filtered too)
	config.Partitions = 0
	config.Segments = 4
	config.Pass = 1
	config.Qmin = 0
	config.Qmax = 100
	config.ShowCompressed = 0
	config.Preprocessing = 0
	config.Autofilter = 0
	config.PartitionLimit = 0
	config.AlphaCompression = 1
	config.AlphaFiltering = 1
	config.AlphaQuality = 100
	config.Lossless = 0
	config.Exact = 0
	config.ImageHint = WEBP_HINT_DEFAULT
	config.EmulateJpegSize = 0
	config.ThreadLevel = 0
	config.LowMemory = 0
	config.NearLossless = 100
	config.UseSharpYUV = 0

	switch preset {
	case WEBP_PRESET_DEFAULT:
	default:
		break
	case WEBP_PRESET_PICTURE:
		config.SnsStrength = 80
		config.FilterSharpness = 4
		config.FilterStrength = 35
		config.Preprocessing &= ^2 // no dithering
	case WEBP_PRESET_PHOTO:
		config.SnsStrength = 80
		config.FilterSharpness = 3
		config.FilterStrength = 30
		config.Preprocessing |= 2
	case WEBP_PRESET_DRAWING:
		config.SnsStrength = 25
		config.FilterSharpness = 6
		config.FilterStrength = 10
	case WEBP_PRESET_ICON:
		config.SnsStrength = 0
		config.FilterStrength = 0  // disable filtering to retain sharpness
		config.Preprocessing &= ^2 // no dithering
	case WEBP_PRESET_TEXT:
		config.SnsStrength = 0
		config.FilterStrength = 0  // disable filtering to retain sharpness
		config.Preprocessing &= ^2 // no dithering
		config.Segments = 2

	}
	return config.Validate()
}

// Returns true if 'config' is non-nil and all configuration parameters are
// within their valid ranges.
func (config *Config) Validate() error {
	if config == nil {
		return errors.New("config is nil")
	}
	if config.Quality < 0 || config.Quality > 100 {
		return errors.New("quality must be between 0 and 100")
	}
	if config.TargetSize < 0 {
		return errors.New("target_size must be non-negative")
	}
	if config.TargetPSNR < 0 {
		return errors.New("target_PSNR must be non-negative")
	}
	if config.Method < 0 || config.Method > 6 {
		return errors.New("method must be between 0 and 6")
	}
	if config.Segments < 1 || config.Segments > 4 {
		return errors.New("segments must be between 1 and 4")
	}
	if config.SnsStrength < 0 || config.SnsStrength > 100 {
		return errors.New("sns_strength must be between 0 and 100")
	}
	if config.FilterStrength < 0 || config.FilterStrength > 100 {
		return errors.New("filter_strength must be between 0 and 100")
	}
	if config.FilterSharpness < 0 || config.FilterSharpness > 7 {
		return errors.New("filter_sharpness must be between 0 and 7")
	}
	if config.FilterType < 0 || config.FilterType > 1 {
		return errors.New("filter_type must be 0 or 1")
	}
	if config.Autofilter < 0 || config.Autofilter > 1 {
		return errors.New("autofilter must be 0 or 1")
	}
	if config.Pass < 1 || config.Pass > 10 {
		return errors.New("pass must be between 1 and 10")
	}
	if config.Qmin < 0 || config.Qmax > 100 || config.Qmin > config.Qmax {
		return errors.New("qmin/qmax must be in [0,100] and qmin <= qmax")
	}
	if config.ShowCompressed < 0 || config.ShowCompressed > 1 {
		return errors.New("show_compressed must be 0 or 1")
	}
	if config.Preprocessing < 0 || config.Preprocessing > 7 {
		return errors.New("preprocessing must be between 0 and 7")
	}
	if config.Partitions < 0 || config.Partitions > 3 {
		return errors.New("partitions must be between 0 and 3")
	}
	if config.PartitionLimit < 0 || config.PartitionLimit > 100 {
		return errors.New("partition_limit must be between 0 and 100")
	}
	if config.AlphaCompression < 0 {
		return errors.New("alpha_compression must be non-negative")
	}
	if config.AlphaFiltering < 0 {
		return errors.New("alpha_filtering must be non-negative")
	}
	if config.AlphaQuality < 0 || config.AlphaQuality > 100 {
		return errors.New("alpha_quality must be between 0 and 100")
	}
	if config.Lossless < 0 || config.Lossless > 1 {
		return errors.New("lossless must be 0 or 1")
	}
	if config.NearLossless < 0 || config.NearLossless > 100 {
		return errors.New("near_lossless must be between 0 and 100")
	}
	if config.ImageHint >= WEBP_HINT_LAST {
		return errors.New("image_hint must be less than WEBP_HINT_LAST")
	}
	if config.EmulateJpegSize < 0 || config.EmulateJpegSize > 1 {
		return errors.New("emulate_jpeg_size must be 0 or 1")
	}
	if config.ThreadLevel < 0 || config.ThreadLevel > 1 {
		return errors.New("thread_level must be 0 or 1")
	}
	if config.LowMemory < 0 || config.LowMemory > 1 {
		return errors.New("low_memory must be 0 or 1")
	}
	if config.Exact < 0 || config.Exact > 1 {
		return errors.New("exact must be 0 or 1")
	}
	if config.UseSharpYUV < 0 || config.UseSharpYUV > 1 {
		return errors.New("use_sharp_yuv must be 0 or 1")
	}

	return nil
}
