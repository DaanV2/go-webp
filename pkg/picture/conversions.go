package picture

func PictureARGBToYUVA(picture *picture.Picture,  colorspace colorspace.CSP ,  dithering float64, use_iterative_conversion int) int {
  if picture == nil { return 0  }
  if (picture.ARGB == nil) {
    return picture.SetEncodingError(picture.VP8_ENC_ERROR_nil_PARAMETER)
  } else if ((colorspace & colorspace.WEBP_CSP_UV_MASK) != colorspace.WEBP_YUV420) {
    return picture.SetEncodingError(picture.VP8_ENC_ERROR_INVALID_CONFIGURATION)
  } else {
    var argb *uint8 = (/* const */ *uint8)picture.ARGB;
    var a *uint8 = argb + CHANNEL_OFFSET(0);
    var r *uint8 = argb + CHANNEL_OFFSET(1);
    var g *uint8 = argb + CHANNEL_OFFSET(2);
    var b *uint8 = argb + CHANNEL_OFFSET(3);

    picture.ColorSpace = colorspace.WEBP_YUV420;
    return ImportYUVAFromRGBA(r, g, b, a, 4, 4 * picture.ARGBStride, dithering, use_iterative_conversion, picture);
  }
}

func WebPPictureARGBToYUVADithered(picture *picture.Picture, colorspace.CSP colorspace, float64 dithering) int {
  return PictureARGBToYUVA(picture, colorspace, dithering, 0);
}

func WebPPictureARGBToYUVA(picture *picture.Picture, colorspace.CSP colorspace) int {
  return PictureARGBToYUVA(picture, colorspace, 0.0, 0);
}

func WebPPictureSharpARGBToYUVA(picture *picture.Picture) int {
  return PictureARGBToYUVA(picture, colorspace.WEBP_YUV420, 0.0, 1);
}

// for backward compatibility
func WebPPictureSmartARGBToYUVA(picture *picture.Picture) int {
  return picture.WebPPictureSharpARGBToYUVA(picture);
}