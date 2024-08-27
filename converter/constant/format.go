package constant

import "github.com/thediveo/enumflag/v2"

type ConversionFormat enumflag.Flag

const (
	WebP ConversionFormat = iota
)

var CommandValue = map[ConversionFormat][]string{
	WebP: {"webp"},
}

var HelpText = enumflag.Help[ConversionFormat]{
	WebP: "WebP Image Format",
}

var DefaultConversion = WebP

func (c ConversionFormat) String() string {
	return CommandValue[c][0]
}

func ListAll() []string {
	var formats []string
	for _, names := range CommandValue {
		formats = append(formats, names[0])
	}
	return formats
}

func FindConversionFormat(format string) ConversionFormat {
	for convFormat, names := range CommandValue {
		for _, name := range names {
			if name == format {
				return convFormat
			}
		}
	}
	return DefaultConversion
}
