package webp

import (
	"bytes"
	"fmt"
	"github.com/belphemur/CBZOptimizer/converter/constant"
	"github.com/belphemur/CBZOptimizer/manga"
	"github.com/oliamb/cutter"
	"golang.org/x/exp/slices"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
)

const webpMaxHeight = 16383

type Converter struct {
	maxHeight  int
	cropHeight int
	isPrepared bool
}

func (converter *Converter) Format() (format constant.ConversionFormat) {
	return constant.WebP
}

func New() *Converter {
	return &Converter{
		//maxHeight: 16383 / 2,
		maxHeight:  4000,
		cropHeight: 2000,
		isPrepared: false,
	}
}

func (converter *Converter) PrepareConverter() error {
	if converter.isPrepared {
		return nil
	}
	err := PrepareEncoder()
	if err != nil {
		return err
	}
	converter.isPrepared = true
	return nil
}

func (converter *Converter) ConvertChapter(chapter *manga.Chapter, quality uint8, split bool, progress func(message string, current uint32, total uint32)) (*manga.Chapter, error) {
	err := converter.PrepareConverter()
	if err != nil {
		return nil, err
	}

	var wgConvertedPages sync.WaitGroup
	maxGoroutines := runtime.NumCPU()

	pagesChan := make(chan *manga.PageContainer, maxGoroutines)
	errChan := make(chan error, maxGoroutines)

	var wgPages sync.WaitGroup
	wgPages.Add(len(chapter.Pages))

	guard := make(chan struct{}, maxGoroutines)
	pagesMutex := sync.Mutex{}
	var pages []*manga.Page
	var totalPages = uint32(len(chapter.Pages))

	go func() {
		for page := range pagesChan {
			guard <- struct{}{} // would block if guard channel is already filled
			go func(pageToConvert *manga.PageContainer) {
				defer wgConvertedPages.Done()
				convertedPage, err := converter.convertPage(pageToConvert, quality)
				if err != nil {
					if convertedPage == nil {
						errChan <- err
						<-guard
						return
					}
					buffer := new(bytes.Buffer)
					err := png.Encode(buffer, convertedPage.Image)
					if err != nil {
						errChan <- err
						<-guard
						return
					}
					convertedPage.Page.Contents = buffer
					convertedPage.Page.Extension = ".png"
					convertedPage.Page.Size = uint64(buffer.Len())
				}
				pagesMutex.Lock()
				pages = append(pages, convertedPage.Page)
				progress(fmt.Sprintf("Converted %d/%d pages to %s format", len(pages), totalPages, converter.Format()), uint32(len(pages)), totalPages)
				pagesMutex.Unlock()
				<-guard
			}(page)
		}
	}()

	for _, page := range chapter.Pages {
		go func(page *manga.Page) {
			defer wgPages.Done()

			splitNeeded, img, format, err := converter.checkPageNeedsSplit(page)
			// Respect choice to split or not
			splitNeeded = split && splitNeeded
			if err != nil {
				errChan <- fmt.Errorf("error checking if page %d of genTestChapter %s needs split: %v", page.Index, chapter.FilePath, err)
				return
			}

			if !splitNeeded {
				wgConvertedPages.Add(1)
				pagesChan <- manga.NewContainer(page, img, format)
				return
			}
			images, err := converter.cropImage(img)
			if err != nil {
				errChan <- fmt.Errorf("error converting page %d of genTestChapter %s to webp: %v", page.Index, chapter.FilePath, err)
				return
			}

			atomic.AddUint32(&totalPages, uint32(len(images)-1))
			for i, img := range images {
				page := &manga.Page{Index: page.Index, IsSplitted: true, SplitPartIndex: uint16(i)}
				wgConvertedPages.Add(1)
				pagesChan <- manga.NewContainer(page, img, "N/A")
			}
		}(page)
	}

	wgPages.Wait()
	wgConvertedPages.Wait()
	close(pagesChan)
	close(errChan)

	var errList []error
	for err := range errChan {
		errList = append(errList, err)
	}

	if len(errList) > 0 {
		return nil, fmt.Errorf("encountered errors: %v", errList)
	}

	slices.SortFunc(pages, func(a, b *manga.Page) int {
		if a.Index == b.Index {
			return int(b.SplitPartIndex - a.SplitPartIndex)
		}
		return int(b.Index - a.Index)
	})
	chapter.Pages = pages

	runtime.GC()

	return chapter, nil
}

func (converter *Converter) cropImage(img image.Image) ([]image.Image, error) {
	bounds := img.Bounds()
	height := bounds.Dy()

	numParts := height / converter.cropHeight
	if height%converter.cropHeight != 0 {
		numParts++
	}

	parts := make([]image.Image, numParts)

	for i := 0; i < numParts; i++ {
		partHeight := converter.cropHeight
		if i == numParts-1 {
			partHeight = height - i*converter.cropHeight
		}

		part, err := cutter.Crop(img, cutter.Config{
			Width:  bounds.Dx(),
			Height: partHeight,
			Anchor: image.Point{Y: i * converter.cropHeight},
			Mode:   cutter.TopLeft,
		})
		if err != nil {
			return nil, fmt.Errorf("error cropping part %d: %v", i+1, err)
		}

		parts[i] = part
	}

	return parts, nil
}

func (converter *Converter) checkPageNeedsSplit(page *manga.Page) (bool, image.Image, string, error) {
	reader := io.Reader(bytes.NewBuffer(page.Contents.Bytes()))
	img, format, err := image.Decode(reader)
	if err != nil {
		return false, nil, format, err
	}

	bounds := img.Bounds()
	height := bounds.Dy()

	if height >= webpMaxHeight {
		return false, img, format, fmt.Errorf("page[%d] height %d exceeds maximum height %d of webp format", page.Index, height, webpMaxHeight)
	}
	return height >= converter.maxHeight, img, format, nil
}

func (converter *Converter) convertPage(container *manga.PageContainer, quality uint8) (*manga.PageContainer, error) {
	if container.Format == "webp" {
		return container, nil
	}
	converted, err := converter.convert(container.Image, uint(quality))
	if err != nil {
		return nil, err
	}
	container.Page.Contents = converted
	container.Page.Extension = ".webp"
	container.Page.Size = uint64(converted.Len())
	return container, nil
}

// convert converts an image to the WebP format. It decodes the image from the input buffer,
// encodes it as a WebP file using the webp.Encode() function, and returns the resulting WebP
// file as a bytes.Buffer.
func (converter *Converter) convert(image image.Image, quality uint) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	err := Encode(&buf, image, quality)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}
