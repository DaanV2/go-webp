package webp

import (
	"errors"
	"image"
	"io"

	"github.com/daanv2/go-webp/pkg/config"
	"github.com/daanv2/go-webp/pkg/libwebp/enc"
)

func Encode(w io.Writer, img image.Image, conf *config.Config) error {
	if conf == nil {
		return errors.New("options is nil")
	}
	if img == nil {
		return errors.New("img is nil")
	}
	if w == nil {
		return errors.New("writer is nil")
	}

	return enc.WebPEncode(conf, img)
}
