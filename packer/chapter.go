package packer

type Chapter struct {
	// FilePath is the path to the chapter's directory.
	FilePath string
	// Pages is a slice of pointers to Page objects.
	Pages []*Page
	// ComicInfo is a string containing information about the chapter.
	ComicInfoXml string
}
