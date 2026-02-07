package alpha

import "image/color"

// Returns true if alpha[] has non-0xff values.
func CheckNonOpaque(/* const */ alpha []color.Alpha, width, height int, x_step int, y_step int) int {
  if alpha == nil { return 0  }
  WebPInitAlphaProcessing();
  if x_step == 1 {
    for i := 0; i < height; i++ {
      if WebPHasAlpha8b(alpha, width) { return 1 }
      alpha += y_step
    }
  } else {
    for i := 0; i < height; i++ {
      if WebPHasAlpha32b(alpha, width) { return 1 }
      alpha += y_step
    }
  }
  return 0;
}

func HasAlpha8b_C(/* const */ src *uint8, length int) int {
  while (length-- > 0) {
    if *src++ != 0xff { return 1  }
  }
  return 0;
}

func HasAlpha32b_C(/* const */ src *uint8, length int) int {
  var x int
  for x = 0; length-- > 0; x += 4 {
    if src[x] != 0xff { return 1  }
  }
  return 0;
}

func AlphaReplace_C(src *uint32, length int, color uint32) {
  var x int
  for x = 0; x < length; x++ {
    if (src[x] >> 24) == 0 { src[x] = color }
  }
}