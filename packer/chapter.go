package packer

import "time"

type Chapter struct {
	// FilePath is the path to the chapter's directory.
	FilePath string
	// Pages is a slice of pointers to Page objects.
	Pages []*Page
	// ComicInfo is a string containing information about the chapter.
	ComicInfoXml string
	// IsConverted is a boolean that indicates whether the chapter has been converted.
	IsConverted bool
	// ConvertedTime is a pointer to a time.Time object that indicates when the chapter was converted. Nil mean not converted.
	ConvertedTime time.Time
}

// SetConverted sets the IsConverted field to true and sets the ConvertedTime field to the current time.
func (chapter *Chapter) SetConverted() {
	chapter.IsConverted = true
	chapter.ConvertedTime = time.Now()
}
