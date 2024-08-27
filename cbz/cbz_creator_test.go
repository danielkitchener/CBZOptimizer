package cbz

import (
	"CBZOptimizer/packer"
	"archive/zip"
	"bytes"
	"os"
	"testing"
)

func TestWriteChapterToCBZ(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name          string
		chapter       *packer.Chapter
		expectedFiles []string
	}{
		{
			name: "Single page, no ComicInfo",
			chapter: &packer.Chapter{
				Pages: []*packer.Page{
					{
						Index:     0,
						Extension: ".jpg",
						Contents:  bytes.NewBuffer([]byte("image data")),
					},
				},
			},
			expectedFiles: []string{"page_0000.jpg"},
		},
		{
			name: "Multiple pages with ComicInfo",
			chapter: &packer.Chapter{
				Pages: []*packer.Page{
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
			expectedFiles: []string{"page_0000.jpg", "page_0001.jpg", "ComicInfo.xml"},
		},
		{
			name: "Split page",
			chapter: &packer.Chapter{
				Pages: []*packer.Page{
					{
						Index:          0,
						Extension:      ".jpg",
						Contents:       bytes.NewBuffer([]byte("split image data")),
						IsSplitted:     true,
						SplitPartIndex: 1,
					},
				},
			},
			expectedFiles: []string{"page_0000-01.jpg"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary file for the .cbz output
			tempFile, err := os.CreateTemp("", "*.cbz")
			if err != nil {
				t.Fatalf("Failed to create temporary file: %v", err)
			}
			defer os.Remove(tempFile.Name())

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
			defer r.Close()

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

			// Check if there are no unexpected files
			if len(filesInArchive) != len(tc.expectedFiles) {
				t.Errorf("Expected %d files, but found %d", len(tc.expectedFiles), len(filesInArchive))
			}
		})
	}
}
