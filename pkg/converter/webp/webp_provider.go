package webp

import (
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"image"
	"io"
)

func PrepareEncoder() error {
	return nil
}
func Encode(w io.Writer, m image.Image, quality uint) error {
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, float32(quality))
	if err != nil {
		return err
	}
	return webp.Encode(w, m, options)
}
