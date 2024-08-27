package cmd

import (
	"github.com/belphemur/CBZOptimizer/cbz"
	"github.com/belphemur/CBZOptimizer/converter"
	"github.com/belphemur/CBZOptimizer/converter/constant"
	"github.com/belphemur/CBZOptimizer/packer"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// MockConverter is a mock implementation of the Converter interface
type MockConverter struct{}

func (m *MockConverter) Format() constant.ConversionFormat {
	return constant.WebP
}

func (m *MockConverter) ConvertChapter(chapter *packer.Chapter, quality uint8, progress func(string)) (*packer.Chapter, error) {
	// Simulate conversion by setting the IsConverted flag
	chapter.IsConverted = true
	chapter.ConvertedTime = time.Now()
	return chapter, nil
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
	defer os.RemoveAll(tempDir) // Clean up the temp directory when done

	// Locate the testdata directory
	testdataDir := filepath.Join("../testdata")
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Fatalf("testdata directory not found")
	}

	// Copy sample CBZ files from testdata to the temporary directory
	err = filepath.Walk(testdataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".cbz") {
			destPath := filepath.Join(tempDir, info.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(destPath, data, info.Mode())
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
	cmd.Flags().BoolP("override", "o", false, "Override the original CBZ files")

	// Execute the command
	err = ConvertCbzCommand(cmd, []string{tempDir})
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	// Verify the results
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), "_converted.cbz") {
			return nil
		}
		t.Logf("CBZ file found: %s", path)

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
		t.Logf("CBZ file [%s] is converted: %s", path, chapter.ConvertedTime)

		return nil
	})
	if err != nil {
		t.Fatalf("Error verifying converted files: %v", err)
	}
}
