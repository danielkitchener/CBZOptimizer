package converter

import (
	"bytes"
	"github.com/belphemur/CBZOptimizer/packer"
	"image"
	"image/jpeg"
	"os"
	"testing"
)

func TestConvertChapter(t *testing.T) {

	testCases := []struct {
		name           string
		genTestChapter func(path string) (*packer.Chapter, error)
	}{
		{
			name:           "All split pages",
			genTestChapter: loadTestChapterSplit,
		},
		{
			name:           "No split pages",
			genTestChapter: loadTestChapterNoSplit,
		},
		{
			name:           "Mix of split and no split pages",
			genTestChapter: loadTestChapterMixSplit,
		},
	}
	// Load test genTestChapter from testdata
	temp, err := os.CreateTemp("", "test_chapter_*.cbz")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)

	}
	defer os.Remove(temp.Name())
	for _, converter := range Available() {
		converter, err := Get(converter)
		if err != nil {
			t.Fatalf("failed to get converter: %v", err)
		}
		t.Run(converter.Format().String(), func(t *testing.T) {
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					chapter, err := tc.genTestChapter(temp.Name())
					if err != nil {
						t.Fatalf("failed to load test genTestChapter: %v", err)
					}

					quality := uint8(80)

					progress := func(msg string) {
						t.Log(msg)
					}

					convertedChapter, err := converter.ConvertChapter(chapter, quality, progress)
					if err != nil {
						t.Fatalf("failed to convert genTestChapter: %v", err)
					}

					if len(convertedChapter.Pages) == 0 {
						t.Fatalf("no pages were converted")
					}

					for _, page := range convertedChapter.Pages {
						if page.Extension != ".webp" {
							t.Errorf("page %d was not converted to webp format", page.Index)
						}
					}
				})
			}
		})
	}
}

func loadTestChapterSplit(path string) (*packer.Chapter, error) {
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

func loadTestChapterNoSplit(path string) (*packer.Chapter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var pages []*packer.Page
	for i := 0; i < 5; i++ { // Assuming there are 5 pages for the test
		img := image.NewRGBA(image.Rect(0, 0, 300, 1000))
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

func loadTestChapterMixSplit(path string) (*packer.Chapter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var pages []*packer.Page
	for i := 0; i < 5; i++ { // Assuming there are 5 pages for the test
		img := image.NewRGBA(image.Rect(0, 0, 300, 1000*(i+1)))
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
