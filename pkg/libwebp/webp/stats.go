package webp

type WebPAuxStats struct {
	coded_size int // final size

	PSNR        [5]float64 // peak-signal-to-noise ratio for Y/U/V/All/Alpha
	block_count [3]int     // number of intra4/intra16/skipped macroblocks

	// approximate number of bytes spent for header
	// and mode-partition #0
	header_bytes [2]int

	// approximate number of bytes spent for
	// DC/AC/uv coefficients for each (0..3) segments.
	residual_bytes [3][4]int
	segment_size   [4]int // number of macroblocks in each segments
	segment_quant  [4]int // quantizer values for each segments
	segment_level  [4]int // filtering strength for each segments [0..63]

	alpha_data_size int // size of the transparency data
	layer_data_size int // size of the enhancement layer data

	// lossless encoder statistics
	// bit0:predictor
	// bit1:cross-color transform
	// bit2:subtract-green
	// bit3:color indexing
	lossless_features          uint32
	histogram_bits             int // number of precision bits of histogram
	transform_bits             int // precision bits for predictor transform
	cache_bits                 int // number of bits for color cache lookup
	palette_size               int // number of color in palette, if used
	lossless_size              int // final lossless size
	lossless_hdr_size          int // lossless header (transform, huffman etc) size
	lossless_data_size         int // lossless image data size
	cross_color_transform_bits int // precision bits for cross-color transform

	pad [1]uint32 // padding for later use
}
