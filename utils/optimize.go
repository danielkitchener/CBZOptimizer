package utils

import (
	"errors"
	"fmt"
	"github.com/belphemur/CBZOptimizer/v2/cbz"
	"github.com/belphemur/CBZOptimizer/v2/converter"
	errors2 "github.com/belphemur/CBZOptimizer/v2/converter/errors"
	"log"
	"strings"
)

type OptimizeOptions struct {
	ChapterConverter converter.Converter
	Path             string
	Quality          uint8
	Override         bool
	Split            bool
}

// Optimize optimizes a CBZ file using the specified converter.
func Optimize(options *OptimizeOptions) error {
	log.Printf("Processing file: %s\n", options.Path)

	// Load the chapter
	chapter, err := cbz.LoadChapter(options.Path)
	if err != nil {
		return fmt.Errorf("failed to load chapter: %v", err)
	}

	if chapter.IsConverted {
		log.Printf("Chapter already converted: %s", options.Path)
		return nil
	}

	// Convert the chapter
	convertedChapter, err := options.ChapterConverter.ConvertChapter(chapter, options.Quality, options.Split, func(msg string, current uint32, total uint32) {
		if current%10 == 0 || current == total {
			log.Printf("[%s] Converting: %d/%d", chapter.FilePath, current, total)
		}
	})
	if err != nil {
		var pageIgnoredError *errors2.PageIgnoredError
		if !errors.As(err, &pageIgnoredError) {
			return fmt.Errorf("failed to convert chapter: %v", err)
		}
	}
	if convertedChapter == nil {
		return fmt.Errorf("failed to convert chapter")
	}

	convertedChapter.SetConverted()

	// Write the converted chapter back to a CBZ file
	outputPath := options.Path
	if !options.Override {
		outputPath = strings.TrimSuffix(options.Path, ".cbz") + "_converted.cbz"
	}
	err = cbz.WriteChapterToCBZ(convertedChapter, outputPath)
	if err != nil {
		return fmt.Errorf("failed to write converted chapter: %v", err)
	}

	log.Printf("Converted file written to: %s\n", outputPath)
	return nil

}
