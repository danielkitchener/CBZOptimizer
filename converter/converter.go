package converter

import (
	"CBZOptimizer/converter/constant"
	"CBZOptimizer/converter/webp"
	"CBZOptimizer/packer"
	"fmt"
	"github.com/samber/lo"
	"strings"
)

type Converter interface {
	// Format of the converter
	Format() (format constant.ConversionFormat)
	ConvertChapter(chapter *packer.Chapter, quality uint8, progress func(string)) (*packer.Chapter, error)
	PrepareConverter() error
}

var converters = map[constant.ConversionFormat]Converter{
	constant.ImageFormatWebP: webp.New(),
}

// Available returns a list of available converters.
func Available() []constant.ConversionFormat {
	return lo.Keys(converters)
}

// Get returns a packer by name.
// If the packer is not available, an error is returned.
func Get(name constant.ConversionFormat) (Converter, error) {
	if packer, ok := converters[name]; ok {
		return packer, nil
	}

	return nil, fmt.Errorf("unkown converter \"%s\", available options are %s", name, strings.Join(lo.Map(Available(), func(item constant.ConversionFormat, index int) string {
		return string(item)
	}), ", "))
}
