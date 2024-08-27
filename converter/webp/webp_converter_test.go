package webp

import (
	"bytes"
	"github.com/belphemur/CBZOptimizer/packer"
	"image"
	"image/jpeg"
	"os"
	"testing"
)

func TestConvertChapter(t *testing.T) {

	// Load test chapter from testdata
	temp, err := os.CreateTemp("", "test_chapter_*.cbz")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)

	}
	defer os.Remove(temp.Name())
	chapter, err := loadTestChapter(temp.Name())
	if err != nil {
		t.Fatalf("failed to load test chapter: %v", err)
	}

	converter := New()
	quality := uint8(80)

	progress := func(msg string) {
		t.Log(msg)
	}

	convertedChapter, err := converter.ConvertChapter(chapter, quality, progress)
	if err != nil {
		t.Fatalf("failed to convert chapter: %v", err)
	}

	if len(convertedChapter.Pages) == 0 {
		t.Fatalf("no pages were converted")
	}

	for _, page := range convertedChapter.Pages {
		if page.Extension != ".webp" {
			t.Errorf("page %d was not converted to webp format", page.Index)
		}
	}
}

func loadTestChapter(path string) (*packer.Chapter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var pages []*packer.Page
	for i := 0; i < 5; i++ { // Assuming there are 5 pages for the test
		img := image.NewRGBA(image.Rect(0, 0, 300, 10000))
		buf := new(bytes.Buffer)
		err := jpeg.Encode(buf, img, nil)
		if err != nil {
			return nil, err
		}
		page := &packer.Page{
			Index:     uint16(i),
			Contents:  buf,
			Extension: ".jpg",
		}
		pages = append(pages, page)
	}

	return &packer.Chapter{
		FilePath: path,
		Pages:    pages,
	}, nil
}
