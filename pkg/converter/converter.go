package converter

import (
	"context"
	"fmt"
	"strings"

	"github.com/dkitchener/CBZOptimizer/v2/internal/manga"
	"github.com/dkitchener/CBZOptimizer/v2/pkg/converter/constant"
	"github.com/dkitchener/CBZOptimizer/v2/pkg/converter/webp"
	"github.com/samber/lo"
)

type Converter interface {
	// Format of the converter
	Format() (format constant.ConversionFormat)
	// ConvertChapter converts a manga chapter to the specified format.
	//
	// Returns partial success where some pages are converted and some are not.
	ConvertChapter(ctx context.Context, chapter *manga.Chapter, quality uint8, lossless bool, split bool, progress func(message string, current uint32, total uint32)) (*manga.Chapter, error)
	PrepareConverter() error
}

var converters = map[constant.ConversionFormat]Converter{
	constant.WebP: webp.New(),
}

// Available returns a list of available converters.
func Available() []constant.ConversionFormat {
	return lo.Keys(converters)
}

// Get returns a packer by name.
// If the packer is not available, an error is returned.
var Get = getConverter

func getConverter(name constant.ConversionFormat) (Converter, error) {
	if converter, ok := converters[name]; ok {
		return converter, nil
	}

	return nil, fmt.Errorf("unkown converter \"%s\", available options are %s", name, strings.Join(lo.Map(Available(), func(item constant.ConversionFormat, index int) string {
		return item.String()
	}), ", "))
}
