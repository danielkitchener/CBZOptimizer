package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/belphemur/CBZOptimizer/v2/internal/cbz"
	"github.com/belphemur/CBZOptimizer/v2/pkg/converter"
	errors2 "github.com/belphemur/CBZOptimizer/v2/pkg/converter/errors"
	"github.com/rs/zerolog/log"
)

type OptimizeOptions struct {
	ChapterConverter converter.Converter
	Path             string
	Quality          uint8
	Override         bool
	Split            bool
}

// Optimize optimizes a CBZ/CBR file using the specified converter.
func Optimize(options *OptimizeOptions) error {
	log.Info().Str("file", options.Path).Msg("Processing file")

	// Load the chapter
	chapter, err := cbz.LoadChapter(options.Path)
	if err != nil {
		return fmt.Errorf("failed to load chapter: %v", err)
	}

	if chapter.IsConverted {
		log.Info().Str("file", options.Path).Msg("Chapter already converted")
		return nil
	}

	// Convert the chapter
	convertedChapter, err := options.ChapterConverter.ConvertChapter(chapter, options.Quality, options.Split, func(msg string, current uint32, total uint32) {
		if current%10 == 0 || current == total {
			log.Info().Str("file", chapter.FilePath).Uint32("current", current).Uint32("total", total).Msg("Converting")
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

	// Determine output path and handle CBR override logic
	outputPath := options.Path
	originalPath := options.Path
	isCbrOverride := false

	if options.Override {
		// For override mode, check if it's a CBR file that needs to be converted to CBZ
		pathLower := strings.ToLower(options.Path)
		if strings.HasSuffix(pathLower, ".cbr") {
			// Convert CBR to CBZ: change extension and mark for deletion
			outputPath = strings.TrimSuffix(options.Path, filepath.Ext(options.Path)) + ".cbz"
			isCbrOverride = true
		}
		// For CBZ files, outputPath remains the same (overwrite)
	} else {
		// Handle both .cbz and .cbr files - strip the extension and add _converted.cbz
		pathLower := strings.ToLower(options.Path)
		if strings.HasSuffix(pathLower, ".cbz") {
			outputPath = strings.TrimSuffix(options.Path, ".cbz") + "_converted.cbz"
		} else if strings.HasSuffix(pathLower, ".cbr") {
			outputPath = strings.TrimSuffix(options.Path, ".cbr") + "_converted.cbz"
		} else {
			// Fallback for other extensions - just add _converted.cbz
			outputPath = options.Path + "_converted.cbz"
		}
	}

	// Write the converted chapter to CBZ file
	err = cbz.WriteChapterToCBZ(convertedChapter, outputPath)
	if err != nil {
		return fmt.Errorf("failed to write converted chapter: %v", err)
	}

	// If we're overriding a CBR file, delete the original CBR after successful write
	if isCbrOverride {
		err = os.Remove(originalPath)
		if err != nil {
			// Log the error but don't fail the operation since conversion succeeded
			log.Warn().Str("file", originalPath).Err(err).Msg("Failed to delete original CBR file")
		} else {
			log.Info().Str("file", originalPath).Msg("Deleted original CBR file")
		}
	}

	log.Info().Str("output", outputPath).Msg("Converted file written")
	return nil

}
