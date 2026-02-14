package yuv

const (
	YUV_FIX  = 16 // fixed-point precision for RGB.YUV
	YUV_HALF = 1 << (YUV_FIX - 1)

	YUV_FIX2  = 6 // fixed-point precision for YUV.RGB
	YUV_MASK2 = (256 << YUV_FIX2) - 1
)