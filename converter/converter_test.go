package converter

import (
	"bytes"
	"github.com/belphemur/CBZOptimizer/converter/constant"
	"github.com/belphemur/CBZOptimizer/manga"
	"golang.org/x/exp/slices"
	"image"
	"image/jpeg"
	"os"
	"testing"
)

func TestConvertChapter(t *testing.T) {

	testCases := []struct {
		name           string
		genTestChapter func(path string) (*manga.Chapter, error)
		split          bool
		expectFailure  []constant.ConversionFormat
	}{
		{
			name:           "All split pages",
			genTestChapter: genBigPages,
			split:          true,
			expectFailure:  []constant.ConversionFormat{},
		},
		{
			name:           "Big Pages, no split",
			genTestChapter: genBigPages,
			split:          false,
			expectFailure:  []constant.ConversionFormat{constant.WebP},
		},
		{
			name:           "No split pages",
			genTestChapter: genSmallPages,
			split:          false,
			expectFailure:  []constant.ConversionFormat{},
		},
		{
			name:           "Mix of split and no split pages",
			genTestChapter: genMixSmallBig,
			split:          true,
			expectFailure:  []constant.ConversionFormat{},
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

					progress := func(msg string, current uint32, total uint32) {
						t.Log(msg)
					}

					convertedChapter, err := converter.ConvertChapter(chapter, quality, tc.split, progress)
					if err != nil {
						if slices.Contains(tc.expectFailure, converter.Format()) {
							t.Logf("Expected failure to convert genTestChapter: %v", err)
							return
						}
						t.Fatalf("failed to convert genTestChapter: %v", err)
					} else if slices.Contains(tc.expectFailure, converter.Format()) {
						t.Fatalf("expected failure to convert genTestChapter didn't happen")
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

func genBigPages(path string) (*manga.Chapter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var pages []*manga.Page
	for i := 0; i < 5; i++ { // Assuming there are 5 pages for the test
		img := image.NewRGBA(image.Rect(0, 0, 300, 17000))
		buf := new(bytes.Buffer)
		err := jpeg.Encode(buf, img, nil)
		if err != nil {
			return nil, err
		}
		page := &manga.Page{
			Index:     uint16(i),
			Contents:  buf,
			Extension: ".jpg",
		}
		pages = append(pages, page)
	}

	return &manga.Chapter{
		FilePath: path,
		Pages:    pages,
	}, nil
}

func genSmallPages(path string) (*manga.Chapter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var pages []*manga.Page
	for i := 0; i < 5; i++ { // Assuming there are 5 pages for the test
		img := image.NewRGBA(image.Rect(0, 0, 300, 1000))
		buf := new(bytes.Buffer)
		err := jpeg.Encode(buf, img, nil)
		if err != nil {
			return nil, err
		}
		page := &manga.Page{
			Index:     uint16(i),
			Contents:  buf,
			Extension: ".jpg",
		}
		pages = append(pages, page)
	}

	return &manga.Chapter{
		FilePath: path,
		Pages:    pages,
	}, nil
}

func genMixSmallBig(path string) (*manga.Chapter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var pages []*manga.Page
	for i := 0; i < 5; i++ { // Assuming there are 5 pages for the test
		img := image.NewRGBA(image.Rect(0, 0, 300, 1000*(i+1)))
		buf := new(bytes.Buffer)
		err := jpeg.Encode(buf, img, nil)
		if err != nil {
			return nil, err
		}
		page := &manga.Page{
			Index:     uint16(i),
			Contents:  buf,
			Extension: ".jpg",
		}
		pages = append(pages, page)
	}

	return &manga.Chapter{
		FilePath: path,
		Pages:    pages,
	}, nil
}
