package cbz

import (
	"archive/zip"
	"fmt"
	"os"
	"time"

	"github.com/dkitchener/CBZOptimizer/v2/internal/manga"
	"github.com/dkitchener/CBZOptimizer/v2/internal/utils/errs"
	"github.com/rs/zerolog/log"
)

func WriteChapterToCBZ(chapter *manga.Chapter, outputFilePath string) error {
	log.Debug().
		Str("chapter_file", chapter.FilePath).
		Str("output_path", outputFilePath).
		Int("page_count", len(chapter.Pages)).
		Bool("is_converted", chapter.IsConverted).
		Msg("Starting CBZ file creation")

	// Create a new ZIP file
	log.Debug().Str("output_path", outputFilePath).Msg("Creating output CBZ file")
	zipFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Error().Str("output_path", outputFilePath).Err(err).Msg("Failed to create CBZ file")
		return fmt.Errorf("failed to create .cbz file: %w", err)
	}
	defer errs.Capture(&err, zipFile.Close, "failed to close .cbz file")

	// Create a new ZIP writer
	log.Debug().Str("output_path", outputFilePath).Msg("Creating ZIP writer")
	zipWriter := zip.NewWriter(zipFile)
	if err != nil {
		log.Error().Str("output_path", outputFilePath).Err(err).Msg("Failed to create ZIP writer")
		return err
	}
	defer errs.Capture(&err, zipWriter.Close, "failed to close .cbz writer")

	// Write each page to the ZIP archive
	log.Debug().Str("output_path", outputFilePath).Int("pages_to_write", len(chapter.Pages)).Msg("Writing pages to CBZ archive")
	for _, page := range chapter.Pages {
		// Construct the file name for the page
		var fileName string
		if page.IsSplitted {
			// Use the format page%03d-%02d for split pages
			fileName = fmt.Sprintf("%04d-%02d%s", page.Index, page.SplitPartIndex, page.Extension)
		} else {
			// Use the format page%03d for non-split pages
			fileName = fmt.Sprintf("%04d%s", page.Index, page.Extension)
		}

		log.Debug().
			Str("output_path", outputFilePath).
			Uint16("page_index", page.Index).
			Bool("is_splitted", page.IsSplitted).
			Uint16("split_part", page.SplitPartIndex).
			Str("filename", fileName).
			Int("size", len(page.Contents.Bytes())).
			Msg("Writing page to CBZ archive")

		// Create a new file in the ZIP archive
		fileWriter, err := zipWriter.CreateHeader(&zip.FileHeader{
			Name:     fileName,
			Method:   zip.Store,
			Modified: time.Now(),
		})
		if err != nil {
			log.Error().Str("output_path", outputFilePath).Str("filename", fileName).Err(err).Msg("Failed to create file in CBZ archive")
			return fmt.Errorf("failed to create file in .cbz: %w", err)
		}

		// Write the page contents to the file
		bytesWritten, err := fileWriter.Write(page.Contents.Bytes())
		if err != nil {
			log.Error().Str("output_path", outputFilePath).Str("filename", fileName).Err(err).Msg("Failed to write page contents")
			return fmt.Errorf("failed to write page contents: %w", err)
		}

		log.Debug().
			Str("output_path", outputFilePath).
			Str("filename", fileName).
			Int("bytes_written", bytesWritten).
			Msg("Page written successfully")
	}

	// Optionally, write the ComicInfo.xml file if present
	if chapter.ComicInfoXml != "" {
		log.Debug().Str("output_path", outputFilePath).Int("xml_size", len(chapter.ComicInfoXml)).Msg("Writing ComicInfo.xml to CBZ archive")
		comicInfoWriter, err := zipWriter.CreateHeader(&zip.FileHeader{
			Name:     "ComicInfo.xml",
			Method:   zip.Deflate,
			Modified: time.Now(),
		})
		if err != nil {
			log.Error().Str("output_path", outputFilePath).Err(err).Msg("Failed to create ComicInfo.xml in CBZ archive")
			return fmt.Errorf("failed to create ComicInfo.xml in .cbz: %w", err)
		}

		bytesWritten, err := comicInfoWriter.Write([]byte(chapter.ComicInfoXml))
		if err != nil {
			log.Error().Str("output_path", outputFilePath).Err(err).Msg("Failed to write ComicInfo.xml contents")
			return fmt.Errorf("failed to write ComicInfo.xml contents: %w", err)
		}
		log.Debug().Str("output_path", outputFilePath).Int("bytes_written", bytesWritten).Msg("ComicInfo.xml written successfully")
	} else {
		log.Debug().Str("output_path", outputFilePath).Msg("No ComicInfo.xml to write")
	}

	if chapter.IsConverted {
		convertedString := fmt.Sprintf("%s\nThis chapter has been converted by CBZOptimizer.", chapter.ConvertedTime)
		log.Debug().Str("output_path", outputFilePath).Str("comment", convertedString).Msg("Setting CBZ comment for converted chapter")
		err = zipWriter.SetComment(convertedString)
		if err != nil {
			log.Error().Str("output_path", outputFilePath).Err(err).Msg("Failed to write CBZ comment")
			return fmt.Errorf("failed to write comment: %w", err)
		}
		log.Debug().Str("output_path", outputFilePath).Msg("CBZ comment set successfully")
	}

	log.Debug().Str("output_path", outputFilePath).Msg("CBZ file creation completed successfully")
	return nil
}
