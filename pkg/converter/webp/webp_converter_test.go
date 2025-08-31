package webp

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"sync"
	"testing"

	_ "golang.org/x/image/webp"

	"github.com/belphemur/CBZOptimizer/v2/internal/manga"
	"github.com/belphemur/CBZOptimizer/v2/pkg/converter/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestImage(width, height int, format string) (image.Image, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create a gradient pattern to ensure we have actual image data
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 255) / width),
				G: uint8((y * 255) / height),
				B: 100,
				A: 255,
			})
		}
	}
	return img, nil
}

func encodeImage(img image.Image, format string) (*bytes.Buffer, string, error) {
	buf := new(bytes.Buffer)

	switch format {
	case "jpeg", "jpg":
		if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return nil, "", err
		}
		return buf, ".jpg", nil
	case "webp":
		PrepareEncoder()
		if err := Encode(buf, img, 80, false); err != nil {
			return nil, "", err
		}
		return buf, ".webp", nil
	case "png":
		fallthrough
	default:
		if err := png.Encode(buf, img); err != nil {
			return nil, "", err
		}
		return buf, ".png", nil
	}
}

func createTestPage(t *testing.T, index int, width, height int, format string) *manga.Page {
	img, err := createTestImage(width, height, format)
	require.NoError(t, err)

	buf, ext, err := encodeImage(img, format)
	require.NoError(t, err)

	return &manga.Page{
		Index:     uint16(index),
		Contents:  buf,
		Extension: ext,
		Size:      uint64(buf.Len()),
	}
}

func validateConvertedImage(t *testing.T, page *manga.Page) {
	require.NotNil(t, page.Contents)
	require.Greater(t, page.Contents.Len(), 0)

	// Try to decode the image
	img, format, err := image.Decode(bytes.NewReader(page.Contents.Bytes()))
	require.NoError(t, err, "Failed to decode converted image")

	if page.Extension == ".webp" {
		assert.Equal(t, "webp", format, "Expected WebP format")
	}

	require.NotNil(t, img)
	bounds := img.Bounds()
	assert.Greater(t, bounds.Dx(), 0, "Image width should be positive")
	assert.Greater(t, bounds.Dy(), 0, "Image height should be positive")
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
			pages:       []*manga.Page{createTestPage(t, 1, 800, 1200, "jpeg")},
			split:       false,
			expectSplit: false,
			numExpected: 1,
		},
		{
			name: "Multiple normal images",
			pages: []*manga.Page{
				createTestPage(t, 1, 800, 1200, "png"),
				createTestPage(t, 2, 800, 1200, "jpeg"),
			},
			split:       false,
			expectSplit: false,
			numExpected: 2,
		},
		{
			name:        "Tall image with split enabled",
			pages:       []*manga.Page{createTestPage(t, 1, 800, 5000, "jpeg")},
			split:       true,
			expectSplit: true,
			numExpected: 3, // Based on cropHeight of 2000
		},
		{
			name:        "Tall image without split",
			pages:       []*manga.Page{createTestPage(t, 1, 800, webpMaxHeight+100, "png")},
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

			convertedChapter, err := converter.ConvertChapter(context.Background(), chapter, 80, false, tt.split, progress)

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

			// Validate all converted images
			for _, page := range convertedChapter.Pages {
				validateConvertedImage(t, page)
			}

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
			page := createTestPage(t, 1, 100, 100, tt.format)
			img, err := createTestImage(100, 100, tt.format)
			require.NoError(t, err)
			container := manga.NewContainer(page, img, tt.format, tt.isToBeConverted)

			converted, err := converter.convertPage(container, 80, false)
			require.NoError(t, err)
			assert.NotNil(t, converted)

			if tt.expectWebP {
				assert.Equal(t, ".webp", converted.Page.Extension)
				validateConvertedImage(t, converted.Page)
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
			page := createTestPage(t, 1, 800, tt.imageHeight, "jpeg")

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

func TestConverter_ConvertChapter_Timeout(t *testing.T) {
	converter := New()
	err := converter.PrepareConverter()
	require.NoError(t, err)

	// Create a test chapter with a few pages
	pages := []*manga.Page{
		createTestPage(t, 1, 800, 1200, "jpeg"),
		createTestPage(t, 2, 800, 1200, "jpeg"),
		createTestPage(t, 3, 800, 1200, "jpeg"),
	}

	chapter := &manga.Chapter{
		FilePath: "/test/chapter.cbz",
		Pages:    pages,
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

	// Test with very short timeout (1 nanosecond)
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()

	convertedChapter, err := converter.ConvertChapter(ctx, chapter, 80, false, false, progress)

	// Should return context error due to timeout
	assert.Error(t, err)
	assert.Nil(t, convertedChapter)
	assert.Equal(t, context.DeadlineExceeded, err)
}
