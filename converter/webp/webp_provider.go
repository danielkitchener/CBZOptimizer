package webp

import (
	"github.com/belphemur/go-webpbin/v2"
	"image"
	"io"
)

const libwebpVersion = "1.4.0"

func PrepareEncoder() error {
	webpbin.SetLibVersion(libwebpVersion)
	container := webpbin.NewCWebP()
	return container.BinWrapper.Run()
}
func Encode(w io.Writer, m image.Image, quality uint) error {
	return webpbin.NewCWebP().
		Quality(quality).
		InputImage(m).
		Output(w).
		Run()
}
