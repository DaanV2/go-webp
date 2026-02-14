package yuv

import "image/color"

var _ color.Color = YUV{}

type YUV struct {
	Y uint8
	U uint8
	V uint8
}

func (c YUV) RGBA() (r, g, b, a uint32) {
	r = uint32(YUVToR(int(c.Y), int(c.V))) * 0x101
	g = uint32(YUVToG(int(c.Y), int(c.U), int(c.V))) * 0x101
	b = uint32(YUVToB(int(c.Y), int(c.U))) * 0x101
	a = 0xffff
	return
}

func (c YUV) YCbCr() (y, cb, cr uint8) {
	return c.Y, c.U, c.V
}

func (c YUV) YCbCrColor() color.YCbCr {
	return color.YCbCr{Y: c.Y, Cb: c.U, Cr: c.V}
}

func (c YUV) RGBAColor() color.RGBA {
	r := uint8(YUVToR(int(c.Y), int(c.V)))
	g := uint8(YUVToG(int(c.Y), int(c.U), int(c.V)))
	b := uint8(YUVToB(int(c.Y), int(c.U)))
	a := uint8(255)
	return color.RGBA{R: r, G: g, B: b, A: a}
}

func (c YUV) Red() uint8 {
	return uint8(YUVToR(int(c.Y), int(c.V)))
}

func (c YUV) Green() uint8 {
	return uint8(YUVToG(int(c.Y), int(c.U), int(c.V)))
}

func (c YUV) Blue() uint8 {
	return uint8(YUVToB(int(c.Y), int(c.U)))
}

func (c YUV) Alpha() uint8 {
	return 255
}

func FromRGBA(r, g, b, a uint8) YUV {
	y := RGBToY(int(r), int(g), int(b), 0)
	u := RGBToU(int(r), int(g), int(b), 0)
	v := RGBToV(int(r), int(g), int(b), 0)
	return YUV{Y: uint8(y), U: uint8(u), V: uint8(v)}
}

func FromRGBAint(r, g, b, a int) YUV {
	y := RGBToY(r, g, b, 0)
	u := RGBToU(r, g, b, 0)
	v := RGBToV(r, g, b, 0)
	return YUV{Y: uint8(y), U: uint8(u), V: uint8(v)}
}

func FromColor(c color.Color) YUV {
	r, g, b, a := c.RGBA()
	return FromRGBA(uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8))
}
