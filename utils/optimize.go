package utils

import (
	"fmt"
	"github.com/belphemur/CBZOptimizer/cbz"
	"github.com/belphemur/CBZOptimizer/converter"
	"log"
	"strings"
)

// Optimize optimizes a CBZ file using the specified converter.
func Optimize(chapterConverter converter.Converter, path string, quality uint8, override bool) error {
	log.Printf("Processing file: %s\n", path)

	// Load the chapter
	chapter, err := cbz.LoadChapter(path)
	if err != nil {
		return fmt.Errorf("failed to load chapter: %v", err)
	}

	if chapter.IsConverted {
		log.Printf("Chapter already converted: %s", path)
		return nil
	}

	// Convert the chapter
	convertedChapter, err := chapterConverter.ConvertChapter(chapter, quality, func(msg string, current uint32, total uint32) {
		if current%10 == 0 || current == total {
			log.Printf("[%s] Converting: %d/%d", chapter.FilePath, current, total)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to convert chapter: %v", err)
	}
	convertedChapter.SetConverted()

	// Write the converted chapter back to a CBZ file
	outputPath := path
	if !override {
		outputPath = strings.TrimSuffix(path, ".cbz") + "_converted.cbz"
	}
	err = cbz.WriteChapterToCBZ(convertedChapter, outputPath)
	if err != nil {
		return fmt.Errorf("failed to write converted chapter: %v", err)
	}

	log.Printf("Converted file written to: %s\n", outputPath)
	return nil

}
