package utils

import (
	"fmt"
	"github.com/belphemur/CBZOptimizer/cbz"
	"github.com/belphemur/CBZOptimizer/converter"
	"strings"
)

// Optimize optimizes a CBZ file using the specified converter.
func Optimize(chapterConverter converter.Converter, path string, quality uint8, override bool) error {
	fmt.Printf("Processing file: %s\n", path)

	// Load the chapter
	chapter, err := cbz.LoadChapter(path)
	if err != nil {
		return fmt.Errorf("failed to load chapter: %v", err)
	}

	if chapter.IsConverted {
		fmt.Printf("Chapter already converted: %s\n", path)
		return nil
	}

	// Convert the chapter
	convertedChapter, err := chapterConverter.ConvertChapter(chapter, quality, func(msg string) {
		fmt.Println(msg)
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

	fmt.Printf("Converted file written to: %s\n", outputPath)
	return nil

}
