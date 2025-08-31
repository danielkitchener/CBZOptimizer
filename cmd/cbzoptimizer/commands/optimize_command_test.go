package commands

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/belphemur/CBZOptimizer/v2/internal/cbz"
	"github.com/belphemur/CBZOptimizer/v2/internal/manga"
	"github.com/belphemur/CBZOptimizer/v2/internal/utils/errs"
	"github.com/belphemur/CBZOptimizer/v2/pkg/converter"
	"github.com/belphemur/CBZOptimizer/v2/pkg/converter/constant"
	"github.com/spf13/cobra"
)

// MockConverter is a mock implementation of the Converter interface
type MockConverter struct{}

func (m *MockConverter) ConvertChapter(ctx context.Context, chapter *manga.Chapter, quality uint8, lossless bool, split bool, progress func(message string, current uint32, total uint32)) (*manga.Chapter, error) {
	chapter.IsConverted = true
	chapter.ConvertedTime = time.Now()
	return chapter, nil
}

func (m *MockConverter) Format() constant.ConversionFormat {
	return constant.WebP
}

func (m *MockConverter) PrepareConverter() error {
	return nil
}

func TestConvertCbzCommand(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_cbz")
	if err != nil {
		log.Fatal(err)
	}
	defer errs.CaptureGeneric(&err, os.RemoveAll, tempDir, "failed to remove temporary directory")

	// Locate the testdata directory
	testdataDir := filepath.Join("../../../testdata")
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Fatalf("testdata directory not found")
	}

	// Copy sample CBZ/CBR files from testdata to the temporary directory
	err = filepath.Walk(testdataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileName := strings.ToLower(info.Name())
			if strings.HasSuffix(fileName, ".cbz") || strings.HasSuffix(fileName, ".cbr") {
				destPath := filepath.Join(tempDir, info.Name())
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				return os.WriteFile(destPath, data, info.Mode())
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to copy sample files: %v", err)
	}

	// Mock the converter.Get function
	originalGet := converter.Get
	converter.Get = func(format constant.ConversionFormat) (converter.Converter, error) {
		return &MockConverter{}, nil
	}
	defer func() { converter.Get = originalGet }()

	// Set up the command
	cmd := &cobra.Command{
		Use: "optimize",
	}
	cmd.Flags().Uint8P("quality", "q", 85, "Quality for conversion (0-100)")
	cmd.Flags().IntP("parallelism", "n", 2, "Number of chapters to convert in parallel")
	cmd.Flags().BoolP("override", "o", false, "Override the original CBZ/CBR files")
	cmd.Flags().BoolP("split", "s", false, "Split long pages into smaller chunks")
	cmd.Flags().DurationP("timeout", "t", 0, "Maximum time allowed for converting a single chapter (e.g., 30s, 5m, 1h). 0 means no timeout")

	// Execute the command
	err = ConvertCbzCommand(cmd, []string{tempDir})
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	// Track expected converted files for verification
	expectedFiles := make(map[string]bool)
	convertedFiles := make(map[string]bool)

	// First pass: identify original files and expected converted filenames
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		fileName := strings.ToLower(info.Name())
		if strings.HasSuffix(fileName, ".cbz") || strings.HasSuffix(fileName, ".cbr") {
			if !strings.Contains(fileName, "_converted") {
				// This is an original file, determine expected converted filename
				baseName := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
				expectedConverted := baseName + "_converted.cbz"
				expectedFiles[expectedConverted] = false // false means not yet found
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Error identifying original files: %v", err)
	}

	// Second pass: verify converted files exist and are properly converted
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		fileName := info.Name()

		// Check if this is a converted file (should only be .cbz, never .cbr)
		if strings.HasSuffix(fileName, "_converted.cbz") {
			convertedFiles[fileName] = true
			expectedFiles[fileName] = true // Mark as found
			t.Logf("Archive file found: %s", path)

			// Load the converted chapter
			chapter, err := cbz.LoadChapter(path)
			if err != nil {
				return err
			}

			// Check if the chapter is marked as converted
			if !chapter.IsConverted {
				t.Errorf("Chapter is not marked as converted: %s", path)
			}

			// Check if the ConvertedTime is set
			if chapter.ConvertedTime.IsZero() {
				t.Errorf("ConvertedTime is not set for chapter: %s", path)
			}
			t.Logf("Archive file [%s] is converted: %s", path, chapter.ConvertedTime)
		} else if strings.HasSuffix(fileName, "_converted.cbr") {
			t.Errorf("Found incorrectly named converted file: %s (should be .cbz, not .cbr)", fileName)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("Error verifying converted files: %v", err)
	}

	// Verify all expected files were found
	for expectedFile, found := range expectedFiles {
		if !found {
			t.Errorf("Expected converted file not found: %s", expectedFile)
		}
	}

	// Log summary
	t.Logf("Found %d converted files", len(convertedFiles))
}
