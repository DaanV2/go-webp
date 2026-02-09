package picture

func PictureARGBToYUVA(picture *picture.Picture,  colorspace colorspace.CSP ,  dithering float64, use_iterative_conversion int) int {
  if picture == nil { return 0  }
  if picture.ARGB == nil {
    return picture.SetEncodingError(picture.VP8_ENC_ERROR_nil_PARAMETER)
  } else if (colorspace & colorspace.WEBP_CSP_UV_MASK) != colorspace.WEBP_YUV420 {
    return picture.SetEncodingError(picture.VP8_ENC_ERROR_INVALID_CONFIGURATION)
  } else {
    // C: const uint8_t* argb = (const uint8_t*)picture->argb
    // C: const uint8_t* a = argb + CHANNEL_OFFSET(0)
    // C: const uint8_t* r = argb + CHANNEL_OFFSET(1)
    // C: const uint8_t* g = argb + CHANNEL_OFFSET(2)
    // C: const uint8_t* b = argb + CHANNEL_OFFSET(3)

    picture.ColorSpace = colorspace.WEBP_YUV420
    return ImportYUVAFromRGBA(r, g, b, a, 4, 4 * picture.ARGBStride, dithering, use_iterative_conversion, picture)
  }
}

func WebPPictureARGBToYUVADithered(picture *picture.Picture, colorspace colorspace.CSP, dithering float64) int {
  return PictureARGBToYUVA(picture, colorspace, dithering, 0)
}

func WebPPictureARGBToYUVA(picture *picture.Picture, colorspace colorspace.CSP) int {
  return PictureARGBToYUVA(picture, colorspace, 0.0, 0)
}

func WebPPictureSharpARGBToYUVA(picture *picture.Picture) int {
  return PictureARGBToYUVA(picture, colorspace.WEBP_YUV420, 0.0, 1)
}

// for backward compatibility
func WebPPictureSmartARGBToYUVA(picture *picture.Picture) int {
  return picture.WebPPictureSharpARGBToYUVA(picture)
}