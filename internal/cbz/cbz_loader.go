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
	"github.com/belphemur/CBZOptimizer/v2/internal/manga"
	"github.com/belphemur/CBZOptimizer/v2/internal/utils/errs"
	"github.com/mholt/archives"
)

func LoadChapter(filePath string) (*manga.Chapter, error) {
	ctx := context.Background()

	chapter := &manga.Chapter{
		FilePath: filePath,
	}

	// First, try to read the comment using zip.OpenReader for CBZ files
	if strings.ToLower(filepath.Ext(filePath)) == ".cbz" {
		r, err := zip.OpenReader(filePath)
		if err == nil {
			defer errs.Capture(&err, r.Close, "failed to close zip reader for comment")

			// Check for comment
			if r.Comment != "" {
				scanner := bufio.NewScanner(strings.NewReader(r.Comment))
				if scanner.Scan() {
					convertedTime := scanner.Text()
					chapter.ConvertedTime, err = dateparse.ParseAny(convertedTime)
					if err == nil {
						chapter.IsConverted = true
					}
				}
			}
		}
		// Continue even if comment reading fails
	}

	// Open the archive using archives library for file operations
	fsys, err := archives.FileSystem(ctx, filePath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive file: %w", err)
	}

	// Walk through all files in the filesystem
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
				// Read the ComicInfo.xml file content
				xmlContent, err := io.ReadAll(file)
				if err != nil {
					return fmt.Errorf("failed to read ComicInfo.xml content: %w", err)
				}
				chapter.ComicInfoXml = string(xmlContent)
			} else if !chapter.IsConverted && ext == ".txt" && fileName == "converted.txt" {
				textContent, err := io.ReadAll(file)
				if err != nil {
					return fmt.Errorf("failed to read converted.txt content: %w", err)
				}
				scanner := bufio.NewScanner(bytes.NewReader(textContent))
				if scanner.Scan() {
					convertedTime := scanner.Text()
					chapter.ConvertedTime, err = dateparse.ParseAny(convertedTime)
					if err != nil {
						return fmt.Errorf("failed to parse converted time: %w", err)
					}
					chapter.IsConverted = true
				}
			} else {
				// Read the file contents for page
				buf := new(bytes.Buffer)
				_, err = io.Copy(buf, file)
				if err != nil {
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
			}
			return nil
		}()
	})

	if err != nil {
		return nil, err
	}

	return chapter, nil
}
