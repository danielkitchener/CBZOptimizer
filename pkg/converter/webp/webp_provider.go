package webp

import (
	"image"
	"io"

	"github.com/danielkitchener/go-webpbin/v2"
)

const libwebpVersion = "1.6.0"

func PrepareEncoder() error {
	webpbin.SetLibVersion(libwebpVersion)
	container := webpbin.NewCWebP()
	return container.BinWrapper.Run()
}
func Encode(w io.Writer, m image.Image, quality uint, lossless bool) error {
	var webp = webpbin.NewCWebP();

	if (lossless) {
		webp.Lossless()
	} else {
		webp.Quality(quality)
	}
	return webp.
		InputImage(m).
		Output(w).
		Run()
}
