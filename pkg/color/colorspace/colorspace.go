package colorspace

// Color spaces.
type CSP int

func (c CSP) Value() int {
	return int(c)
}

const (
	// chroma sampling
	WEBP_YUV420        CSP = 0 // 4:2:0
	WEBP_YUV420A       CSP = 4 // alpha channel variant
	WEBP_CSP_UV_MASK   CSP = 3 // bit-mask to get the UV sampling factors
	WEBP_CSP_ALPHA_BIT CSP = 4 // bit that is set if alpha is present
)
