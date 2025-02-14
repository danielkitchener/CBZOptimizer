package manga

import "bytes"

type Page struct {
	// Index of the page in the chapter.
	Index uint16 `json:"index" jsonschema:"description=Index of the page in the chapter."`
	// Extension of the page image.
	Extension string `json:"extension" jsonschema:"description=Extension of the page image."`
	// Size of the page in bytes
	Size uint64 `json:"-"`
	// Contents of the page
	Contents *bytes.Buffer `json:"-"`
	// IsSplitted tell us if the page was cropped to multiple pieces
	IsSplitted bool `json:"is_cropped" jsonschema:"description=Was this page cropped."`
	// SplitPartIndex represent the index of the crop if the page was cropped
	SplitPartIndex uint16 `json:"crop_part_index" jsonschema:"description=Index of the crop if the image was cropped."`
}
