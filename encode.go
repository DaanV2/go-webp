package webp

import (
	"errors"
	"image"
	"io"

	"github.com/daanv2/go-webp/pkg/config"
	"github.com/daanv2/go-webp/pkg/encoding/lossless"
	"github.com/daanv2/go-webp/pkg/encoding/lossy"
)

func Encode(w io.Writer, img image.Image, options *config.Config) error {
	if options == nil {
		return errors.New("options is nil")
	}
	if img == nil {
		return errors.New("img is nil")
	}
	if w == nil {
		return errors.New("writer is nil")
	}

	switch options.LossLess {
	case 1: // Lossless
		return lossless.Encode(w, img, options)
	case 0: // Lossy
		return lossy.Encode(w, img, options)
	default:
		return errors.New("invalid lossless option, must be 0 (lossy) or 1 (lossless)")
	}
}
