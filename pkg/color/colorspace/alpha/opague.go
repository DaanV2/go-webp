package alpha

import "image/color"

// Returns true if alpha[] has non-0xff values.
func CheckNonOpaque( /* const */ alpha []color.Alpha) bool {
	for _, i := range alpha {
		if uint8(i.A) != 0xff {
			return true
		}
	}

	return false
}

// Returns true if alpha[] has non-0xff values.
func CheckNonOpaque2[T color.Color]( /* const */ clrs []T) bool {
	for _, i := range clrs {
		_, _, _, a := i.RGBA()
		if uint8(a) != 0xff {
			return true
		}
	}

	return false
}

func WebPHasAlpha8b(src []color.Alpha) bool {
	for _, i := range src {
		if uint8(i.A) != 0xff {
			return true
		}
	}
	return false
}

func WebPAlphaReplace(src []color.Alpha, length int, color uint8) {
	for x := range length {
		if src[x].A == 0 {
			src[x].A = color
		}
	}
}
