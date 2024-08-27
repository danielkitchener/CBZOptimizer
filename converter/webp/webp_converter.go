package webp

import (
	"CBZOptimizer/converter/constant"
	packer2 "CBZOptimizer/packer"
	"bytes"
	"fmt"
	"github.com/oliamb/cutter"
	"golang.org/x/exp/slices"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
)

type Converter struct {
	maxHeight  int
	cropHeight int
}

func (converter *Converter) Format() (format constant.ConversionFormat) {
	return constant.ImageFormatWebP
}

func New() *Converter {
	return &Converter{
		//maxHeight: 16383 / 2,
		maxHeight:  4000,
		cropHeight: 2000,
	}
}

func (converter *Converter) ConvertChapter(chapter *packer2.Chapter, quality uint8, progress func(string)) (*packer2.Chapter, error) {
	err := PrepareEncoder()
	if err != nil {
		return nil, err
	}

	var wgConvertedPages sync.WaitGroup
	maxGoroutines := runtime.NumCPU()

	pagesChan := make(chan *packer2.PageContainer, maxGoroutines)

	var wgPages sync.WaitGroup
	wgPages.Add(len(chapter.Pages))

	guard := make(chan struct{}, maxGoroutines)
	pagesMutex := sync.Mutex{}
	var pages []*packer2.Page
	var totalPages = uint32(len(chapter.Pages))

	go func() {
		for page := range pagesChan {
			guard <- struct{}{} // would block if guard channel is already filled
			go func(pageToConvert *packer2.PageContainer) {
				defer wgConvertedPages.Done()
				convertedPage, err := converter.convertPage(pageToConvert, quality)
				if err != nil {
					buffer := new(bytes.Buffer)
					err := png.Encode(buffer, convertedPage.Image)
					if err != nil {
						<-guard
						return
					}
					convertedPage.Page.Contents = buffer
					convertedPage.Page.Extension = ".png"
					convertedPage.Page.Size = uint64(buffer.Len())
				}
				pagesMutex.Lock()
				pages = append(pages, convertedPage.Page)
				progress(fmt.Sprintf("Converted %d/%d pages to %s format", len(pages), totalPages, converter.Format()))
				pagesMutex.Unlock()
				<-guard
			}(page)

		}
	}()

	for _, page := range chapter.Pages {
		go func(page *packer2.Page) {
			defer wgPages.Done()

			splitNeeded, img, format, err := converter.checkPageNeedsSplit(page)
			if err != nil {
				log.Fatalf("error checking if page %d d of chapter %s  needs split: %v", page.Index, chapter.FilePath, err)
				return
			}

			if !splitNeeded {
				wgConvertedPages.Add(1)
				pagesChan <- packer2.NewContainer(page, img, format)
				return
			}
			images, err := converter.cropImage(img)
			if err != nil {
				log.Fatalf("error converting page %d of chapter %s to webp: %v", page.Index, chapter.FilePath, err)
				return
			}

			atomic.AddUint32(&totalPages, uint32(len(images)-1))
			for i, img := range images {
				page := &packer2.Page{Index: page.Index, IsSplitted: true, SplitPartIndex: uint16(i)}
				wgConvertedPages.Add(1)
				pagesChan <- packer2.NewContainer(page, img, "N/A")
			}
		}(page)

	}

	wgPages.Wait()
	wgConvertedPages.Wait()
	close(pagesChan)

	slices.SortFunc(pages, func(a, b *packer2.Page) int {
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

func (converter *Converter) checkPageNeedsSplit(page *packer2.Page) (bool, image.Image, string, error) {
	reader := io.Reader(bytes.NewBuffer(page.Contents.Bytes()))
	img, format, err := image.Decode(reader)
	if err != nil {
		return false, nil, format, err
	}

	bounds := img.Bounds()
	height := bounds.Dy()

	return height >= converter.maxHeight, img, format, nil
}

func (converter *Converter) convertPage(container *packer2.PageContainer, quality uint8) (*packer2.PageContainer, error) {
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

// convert converts an image to the ImageFormatWebP format. It decodes the image from the input buffer,
// encodes it as a ImageFormatWebP file using the webp.Encode() function, and returns the resulting ImageFormatWebP
// file as a bytes.Buffer.
func (converter *Converter) convert(image image.Image, quality uint) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	err := Encode(&buf, image, quality)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}
