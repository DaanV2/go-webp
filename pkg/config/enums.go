package config

// Image characteristics hint for the underlying encoder.
type ImageHint int

const (
	WEBP_HINT_DEFAULT ImageHint = iota // default preset.
	WEBP_HINT_PICTURE                  // digital picture, like portrait, inner shot
	WEBP_HINT_PHOTO                    // outdoor photograph, with natural lighting
	WEBP_HINT_GRAPH                    // Discrete tone image (graph, map-tile etc).
	WEBP_HINT_LAST
)

// Enumerate some predefined settings for config.Config, depending on the type
// of source picture. These presets are used when calling WebPConfigPreset().
type Preset int

const (
	WEBP_PRESET_DEFAULT Preset = iota // default preset.
	WEBP_PRESET_PICTURE               // digital picture, like portrait, inner shot
	WEBP_PRESET_PHOTO                 // outdoor photograph, with natural lighting
	WEBP_PRESET_DRAWING               // hand or line drawing, with high-contrast details
	WEBP_PRESET_ICON                  // small-sized colorful images
	WEBP_PRESET_TEXT                  // text-like
)
