package cbz

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/araddon/dateparse"
	"github.com/dkitchener/CBZOptimizer/v2/internal/manga"
	"github.com/dkitchener/CBZOptimizer/v2/internal/utils/errs"
	"github.com/mholt/archives"
	"github.com/rs/zerolog/log"
)

func LoadChapter(filePath string) (*manga.Chapter, error) {
	log.Debug().Str("file_path", filePath).Msg("Starting chapter loading")

	ctx := context.Background()

	chapter := &manga.Chapter{
		FilePath: filePath,
	}

	// First, try to read the comment using zip.OpenReader for CBZ files
	if strings.ToLower(filepath.Ext(filePath)) == ".cbz" {
		log.Debug().Str("file_path", filePath).Msg("Checking CBZ comment for conversion status")
		r, err := zip.OpenReader(filePath)
		if err == nil {
			defer errs.Capture(&err, r.Close, "failed to close zip reader for comment")

			// Check for comment
			if r.Comment != "" {
				log.Debug().Str("file_path", filePath).Str("comment", r.Comment).Msg("Found CBZ comment")
				scanner := bufio.NewScanner(strings.NewReader(r.Comment))
				if scanner.Scan() {
					convertedTime := scanner.Text()
					log.Debug().Str("file_path", filePath).Str("converted_time", convertedTime).Msg("Parsing conversion timestamp")
					chapter.ConvertedTime, err = dateparse.ParseAny(convertedTime)
					if err == nil {
						chapter.IsConverted = true
						log.Debug().Str("file_path", filePath).Time("converted_time", chapter.ConvertedTime).Msg("Chapter marked as previously converted")
					} else {
						log.Debug().Str("file_path", filePath).Err(err).Msg("Failed to parse conversion timestamp")
					}
				}
			} else {
				log.Debug().Str("file_path", filePath).Msg("No CBZ comment found")
			}
		} else {
			log.Debug().Str("file_path", filePath).Err(err).Msg("Failed to open CBZ file for comment reading")
		}
		// Continue even if comment reading fails
	}

	// Open the archive using archives library for file operations
	log.Debug().Str("file_path", filePath).Msg("Opening archive file system")
	fsys, err := archives.FileSystem(ctx, filePath, nil)
	if err != nil {
		log.Error().Str("file_path", filePath).Err(err).Msg("Failed to open archive file system")
		return nil, fmt.Errorf("failed to open archive file: %w", err)
	}

	// Walk through all files in the filesystem
	log.Debug().Str("file_path", filePath).Msg("Starting filesystem walk")
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		return func() error {
			// Open the file
			file, err := fsys.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}
			defer errs.Capture(&err, file.Close, fmt.Sprintf("failed to close file %s", path))

			// Determine the file extension
			ext := strings.ToLower(filepath.Ext(path))
			fileName := strings.ToLower(filepath.Base(path))

			if ext == ".xml" && fileName == "comicinfo.xml" {
				log.Debug().Str("file_path", filePath).Str("archive_file", path).Msg("Found ComicInfo.xml")
				// Read the ComicInfo.xml file content
				xmlContent, err := io.ReadAll(file)
				if err != nil {
					log.Error().Str("file_path", filePath).Str("archive_file", path).Err(err).Msg("Failed to read ComicInfo.xml")
					return fmt.Errorf("failed to read ComicInfo.xml content: %w", err)
				}
				chapter.ComicInfoXml = string(xmlContent)
				log.Debug().Str("file_path", filePath).Int("xml_size", len(xmlContent)).Msg("ComicInfo.xml loaded")
			} else if !chapter.IsConverted && ext == ".txt" && fileName == "converted.txt" {
				log.Debug().Str("file_path", filePath).Str("archive_file", path).Msg("Found converted.txt")
				textContent, err := io.ReadAll(file)
				if err != nil {
					log.Error().Str("file_path", filePath).Str("archive_file", path).Err(err).Msg("Failed to read converted.txt")
					return fmt.Errorf("failed to read converted.txt content: %w", err)
				}
				scanner := bufio.NewScanner(bytes.NewReader(textContent))
				if scanner.Scan() {
					convertedTime := scanner.Text()
					log.Debug().Str("file_path", filePath).Str("converted_time", convertedTime).Msg("Parsing converted.txt timestamp")
					chapter.ConvertedTime, err = dateparse.ParseAny(convertedTime)
					if err != nil {
						log.Error().Str("file_path", filePath).Err(err).Msg("Failed to parse converted time from converted.txt")
						return fmt.Errorf("failed to parse converted time: %w", err)
					}
					chapter.IsConverted = true
					log.Debug().Str("file_path", filePath).Time("converted_time", chapter.ConvertedTime).Msg("Chapter marked as converted from converted.txt")
				}
			} else {
				// Read the file contents for page
				log.Debug().Str("file_path", filePath).Str("archive_file", path).Str("extension", ext).Msg("Processing page file")
				buf := new(bytes.Buffer)
				bytesCopied, err := io.Copy(buf, file)
				if err != nil {
					log.Error().Str("file_path", filePath).Str("archive_file", path).Err(err).Msg("Failed to read page file contents")
					return fmt.Errorf("failed to read file contents: %w", err)
				}

				// Create a new Page object
				page := &manga.Page{
					Index:      uint16(len(chapter.Pages)), // Simple index based on order
					Extension:  ext,
					Size:       uint64(buf.Len()),
					Contents:   buf,
					IsSplitted: false,
				}

				// Add the page to the chapter
				chapter.Pages = append(chapter.Pages, page)
				log.Debug().
					Str("file_path", filePath).
					Str("archive_file", path).
					Uint16("page_index", page.Index).
					Int64("bytes_read", bytesCopied).
					Msg("Page loaded successfully")
			}
			return nil
		}()
	})

	if err != nil {
		log.Error().Str("file_path", filePath).Err(err).Msg("Failed during filesystem walk")
		return nil, err
	}

	log.Debug().
		Str("file_path", filePath).
		Int("pages_loaded", len(chapter.Pages)).
		Bool("is_converted", chapter.IsConverted).
		Bool("has_comic_info", chapter.ComicInfoXml != "").
		Msg("Chapter loading completed successfully")

	return chapter, nil
}
