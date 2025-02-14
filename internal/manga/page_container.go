package manga

import "image"

// PageContainer is a struct that holds a manga page, its image, and the image format.
type PageContainer struct {
	// Page is a pointer to a manga page object.
	Page *Page
	// Image is the decoded image of the manga page.
	Image image.Image
	// Format is a string representing the format of the image (e.g., "png", "jpeg", "webp").
	Format string
	// IsToBeConverted is a boolean flag indicating whether the image needs to be converted to another format.
	IsToBeConverted bool
}

func NewContainer(Page *Page, img image.Image, format string, isToBeConverted bool) *PageContainer {
	return &PageContainer{Page: Page, Image: img, Format: format, IsToBeConverted: isToBeConverted}
}
