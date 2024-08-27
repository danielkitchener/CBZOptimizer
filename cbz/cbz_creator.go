package cbz

import (
	"CBZOptimizer/packer"
	"archive/zip"
	"fmt"
	"os"
)

func WriteChapterToCBZ(chapter *packer.Chapter, outputFilePath string) error {
	// Create a new ZIP file
	zipFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create .cbz file: %w", err)
	}
	defer zipFile.Close()

	// Create a new ZIP writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Write each page to the ZIP archive
	for _, page := range chapter.Pages {
		// Construct the file name for the page
		var fileName string
		if page.IsSplitted {
			// Use the format page%03d-%02d for split pages
			fileName = fmt.Sprintf("page_%04d-%02d%s", page.Index, page.SplitPartIndex, page.Extension)
		} else {
			// Use the format page%03d for non-split pages
			fileName = fmt.Sprintf("page_%04d%s", page.Index, page.Extension)
		}

		// Create a new file in the ZIP archive
		fileWriter, err := zipWriter.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create file in .cbz: %w", err)
		}

		// Write the page contents to the file
		_, err = fileWriter.Write(page.Contents.Bytes())
		if err != nil {
			return fmt.Errorf("failed to write page contents: %w", err)
		}
	}

	// Optionally, write the ComicInfo.xml file if present
	if chapter.ComicInfoXml != "" {
		comicInfoWriter, err := zipWriter.Create("ComicInfo.xml")
		if err != nil {
			return fmt.Errorf("failed to create ComicInfo.xml in .cbz: %w", err)
		}

		_, err = comicInfoWriter.Write([]byte(chapter.ComicInfoXml))
		if err != nil {
			return fmt.Errorf("failed to write ComicInfo.xml contents: %w", err)
		}
	}

	if chapter.IsConverted {
		convertedWriter, err := zipWriter.Create("Converted.txt")
		if err != nil {
			return fmt.Errorf("failed to create Converted.txt in .cbz: %w", err)
		}

		_, err = convertedWriter.Write([]byte(fmt.Sprintf("%s\nThis chapter has been converted by CBZOptimizer.", chapter.ConvertedTime)))
		if err != nil {
			return fmt.Errorf("failed to write Converted.txt contents: %w", err)
		}
	}

	return nil
}
