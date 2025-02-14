package webp

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"sync"
	"testing"

	"github.com/belphemur/CBZOptimizer/v2/internal/manga"
	"github.com/belphemur/CBZOptimizer/v2/pkg/converter/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	return img
}

func createTestPage(t *testing.T, index int, width, height int) *manga.Page {
	img := createTestImage(width, height)
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	require.NoError(t, err)

	return &manga.Page{
		Index:     uint16(index),
		Contents:  &buf,
		Extension: ".png",
		Size:      uint64(buf.Len()),
	}
}

// TestConverter_ConvertChapter tests the ConvertChapter method of the WebP converter.
// It verifies various scenarios including:
// - Converting single normal images
// - Converting multiple normal images
// - Converting tall images with split enabled
// - Handling tall images that exceed maximum height
//
// For each test case it validates:
// - Proper error handling
// - Expected number of output pages
// - Correct page ordering
// - Split page handling and indexing
// - Progress callback behavior
//
// The test uses different image dimensions and split settings to ensure
// the converter handles all cases correctly while maintaining proper
// progress reporting and page ordering.
func TestConverter_ConvertChapter(t *testing.T) {
	tests := []struct {
		name        string
		pages       []*manga.Page
		split       bool
		expectSplit bool
		expectError bool
		numExpected int
	}{
		{
			name:        "Single normal image",
			pages:       []*manga.Page{createTestPage(t, 1, 800, 1200)},
			split:       false,
			expectSplit: false,
			numExpected: 1,
		},
		{
			name: "Multiple normal images",
			pages: []*manga.Page{
				createTestPage(t, 1, 800, 1200),
				createTestPage(t, 2, 800, 1200),
			},
			split:       false,
			expectSplit: false,
			numExpected: 2,
		},
		{
			name:        "Tall image with split enabled",
			pages:       []*manga.Page{createTestPage(t, 1, 800, 5000)},
			split:       true,
			expectSplit: true,
			numExpected: 3, // Based on cropHeight of 2000
		},
		{
			name:        "Tall image without split",
			pages:       []*manga.Page{createTestPage(t, 1, 800, webpMaxHeight+100)},
			split:       false,
			expectError: true,
			numExpected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := New()
			err := converter.PrepareConverter()
			require.NoError(t, err)

			chapter := &manga.Chapter{
				Pages: tt.pages,
			}

			var progressMutex sync.Mutex
			var lastProgress uint32
			progress := func(message string, current uint32, total uint32) {
				progressMutex.Lock()
				defer progressMutex.Unlock()
				assert.GreaterOrEqual(t, current, lastProgress, "Progress should never decrease")
				lastProgress = current
				assert.LessOrEqual(t, current, total, "Current progress should not exceed total")
			}

			convertedChapter, err := converter.ConvertChapter(chapter, 80, tt.split, progress)

			if tt.expectError {
				assert.Error(t, err)
				if convertedChapter != nil {
					assert.LessOrEqual(t, len(convertedChapter.Pages), tt.numExpected)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, convertedChapter)
			assert.Len(t, convertedChapter.Pages, tt.numExpected)

			// Verify page order
			for i := 1; i < len(convertedChapter.Pages); i++ {
				prevPage := convertedChapter.Pages[i-1]
				currPage := convertedChapter.Pages[i]

				if prevPage.Index == currPage.Index {
					assert.Less(t, prevPage.SplitPartIndex, currPage.SplitPartIndex,
						"Split parts should be in ascending order for page %d", prevPage.Index)
				} else {
					assert.Less(t, prevPage.Index, currPage.Index,
						"Pages should be in ascending order")
				}
			}

			if tt.expectSplit {
				splitFound := false
				for _, page := range convertedChapter.Pages {
					if page.IsSplitted {
						splitFound = true
						break
					}
				}
				assert.True(t, splitFound, "Expected to find at least one split page")
			}
		})
	}
}

func TestConverter_convertPage(t *testing.T) {
	converter := New()
	err := converter.PrepareConverter()
	require.NoError(t, err)

	tests := []struct {
		name            string
		format          string
		isToBeConverted bool
		expectWebP      bool
	}{
		{
			name:            "Convert PNG to WebP",
			format:          "png",
			isToBeConverted: true,
			expectWebP:      true,
		},
		{
			name:            "Already WebP",
			format:          "webp",
			isToBeConverted: true,
			expectWebP:      true,
		},
		{
			name:            "Skip conversion",
			format:          "png",
			isToBeConverted: false,
			expectWebP:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := createTestPage(t, 1, 100, 100)
			container := manga.NewContainer(page, createTestImage(100, 100), tt.format, tt.isToBeConverted)
			defer container.Close()

			converted, err := converter.convertPage(container, 80)
			require.NoError(t, err)
			assert.NotNil(t, converted)

			if tt.expectWebP {
				assert.Equal(t, ".webp", converted.Page.Extension)
			} else {
				assert.NotEqual(t, ".webp", converted.Page.Extension)
			}
		})
	}
}

func TestConverter_checkPageNeedsSplit(t *testing.T) {
	converter := New()

	tests := []struct {
		name        string
		imageHeight int
		split       bool
		expectSplit bool
		expectError bool
	}{
		{
			name:        "Normal height",
			imageHeight: 1000,
			split:       true,
			expectSplit: false,
		},
		{
			name:        "Height exceeds max with split enabled",
			imageHeight: 5000,
			split:       true,
			expectSplit: true,
		},
		{
			name:        "Height exceeds webp max without split",
			imageHeight: webpMaxHeight + 100,
			split:       false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := createTestPage(t, 1, 800, tt.imageHeight)

			needsSplit, img, format, err := converter.checkPageNeedsSplit(page, tt.split)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, img)
			assert.NotEmpty(t, format)
			assert.Equal(t, tt.expectSplit, needsSplit)
		})
	}
}

func TestConverter_Format(t *testing.T) {
	converter := New()
	assert.Equal(t, constant.WebP, converter.Format())
}
