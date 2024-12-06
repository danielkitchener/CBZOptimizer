package cbz

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/belphemur/CBZOptimizer/v2/manga"
	"github.com/belphemur/CBZOptimizer/v2/utils/errs"
	"os"
	"testing"
	"time"
)

func TestWriteChapterToCBZ(t *testing.T) {
	currentTime := time.Now()

	// Define test cases
	testCases := []struct {
		name            string
		chapter         *manga.Chapter
		expectedFiles   []string
		expectedComment string
	}{
		//test case where there is only one page and ComicInfo and the chapter is converted
		{
			name: "Single page, ComicInfo, converted",
			chapter: &manga.Chapter{
				Pages: []*manga.Page{
					{
						Index:     0,
						Extension: ".jpg",
						Contents:  bytes.NewBuffer([]byte("image data")),
					},
				},
				ComicInfoXml:  "<Series>Boundless Necromancer</Series>",
				IsConverted:   true,
				ConvertedTime: currentTime,
			},
			expectedFiles:   []string{"0000.jpg", "ComicInfo.xml"},
			expectedComment: fmt.Sprintf("%s\nThis chapter has been converted by CBZOptimizer.", currentTime),
		},
		//test case where there is only one page and no
		{
			name: "Single page, no ComicInfo",
			chapter: &manga.Chapter{
				Pages: []*manga.Page{
					{
						Index:     0,
						Extension: ".jpg",
						Contents:  bytes.NewBuffer([]byte("image data")),
					},
				},
			},
			expectedFiles: []string{"0000.jpg"},
		},
		{
			name: "Multiple pages with ComicInfo",
			chapter: &manga.Chapter{
				Pages: []*manga.Page{
					{
						Index:     0,
						Extension: ".jpg",
						Contents:  bytes.NewBuffer([]byte("image data 1")),
					},
					{
						Index:     1,
						Extension: ".jpg",
						Contents:  bytes.NewBuffer([]byte("image data 2")),
					},
				},
				ComicInfoXml: "<Series>Boundless Necromancer</Series>",
			},
			expectedFiles: []string{"0000.jpg", "0001.jpg", "ComicInfo.xml"},
		},
		{
			name: "Split page",
			chapter: &manga.Chapter{
				Pages: []*manga.Page{
					{
						Index:          0,
						Extension:      ".jpg",
						Contents:       bytes.NewBuffer([]byte("split image data")),
						IsSplitted:     true,
						SplitPartIndex: 1,
					},
				},
			},
			expectedFiles: []string{"0000-01.jpg"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary file for the .cbz output
			tempFile, err := os.CreateTemp("", "*.cbz")
			if err != nil {
				t.Fatalf("Failed to create temporary file: %v", err)
			}
			defer errs.CaptureGeneric(&err, os.Remove, tempFile.Name(), "failed to remove temporary file")

			// Write the chapter to the .cbz file
			err = WriteChapterToCBZ(tc.chapter, tempFile.Name())
			if err != nil {
				t.Fatalf("Failed to write chapter to CBZ: %v", err)
			}

			// Open the .cbz file as a zip archive
			r, err := zip.OpenReader(tempFile.Name())
			if err != nil {
				t.Fatalf("Failed to open CBZ file: %v", err)
			}
			defer errs.Capture(&err, r.Close, "failed to close CBZ file")

			// Collect the names of the files in the archive
			var filesInArchive []string
			for _, f := range r.File {
				filesInArchive = append(filesInArchive, f.Name)
			}

			// Check if all expected files are present
			for _, expectedFile := range tc.expectedFiles {
				found := false
				for _, actualFile := range filesInArchive {
					if actualFile == expectedFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected file %s not found in archive", expectedFile)
				}
			}

			if tc.expectedComment != "" && r.Comment != tc.expectedComment {
				t.Errorf("Expected comment %s, but found %s", tc.expectedComment, r.Comment)
			}

			// Check if there are no unexpected files
			if len(filesInArchive) != len(tc.expectedFiles) {
				t.Errorf("Expected %d files, but found %d", len(tc.expectedFiles), len(filesInArchive))
			}
		})
	}
}
