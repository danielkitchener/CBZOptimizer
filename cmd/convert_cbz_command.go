package cmd

import (
	"CBZOptimizer/cbz"
	"CBZOptimizer/converter"
	"CBZOptimizer/converter/constant"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var converterType constant.ConversionFormat

func init() {
	command := &cobra.Command{
		Use:   "optimize [folder]",
		Short: "Optimize all CBZ files in a folder recursively",
		Long:  "Optimize all CBZ files in a folder recursively.\nIt will take all the different pages in the CBZ files and convert them to the given format.\nThe original CBZ files will be kept intact depending if you choose to override or not.",
		RunE:  ConvertCbzCommand,
		Args:  cobra.ExactArgs(1),
	}
	formatFlag := enumflag.New(&converterType, "format", constant.CommandValue, enumflag.EnumCaseInsensitive)
	_ = formatFlag.RegisterCompletion(command, "format", constant.HelpText)

	command.Flags().Uint8P("quality", "q", 85, "Quality for conversion (0-100)")
	command.Flags().IntP("parallelism", "n", 2, "Number of chapters to convert in parallel")
	command.Flags().BoolP("override", "o", false, "Override the original CBZ files")
	command.PersistentFlags().VarP(
		formatFlag,
		"format", "f",
		fmt.Sprintf("Format to convert the images to: %s", constant.ListAll()))
	command.PersistentFlags().Lookup("format").NoOptDefVal = constant.DefaultConversion.String()

	AddCommand(command)
}

// isValidFolder checks if the provided path is a valid directory
func isValidFolder(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func ConvertCbzCommand(cmd *cobra.Command, args []string) error {
	path := args[0]
	if path == "" {
		return fmt.Errorf("path is required")
	}

	if !isValidFolder(path) {
		return fmt.Errorf("the path needs to be a folder")
	}

	quality, err := cmd.Flags().GetUint8("quality")
	if err != nil {
		return fmt.Errorf("invalid quality value")
	}

	override, err := cmd.Flags().GetBool("override")
	if err != nil {
		return fmt.Errorf("invalid quality value")
	}

	parallelism, err := cmd.Flags().GetInt("parallelism")
	if err != nil || parallelism < 1 {
		return fmt.Errorf("invalid parallelism value")
	}

	chapterConverter, err := converter.Get(converterType)
	if err != nil {
		return fmt.Errorf("failed to get chapterConverter: %v", err)
	}

	err = chapterConverter.PrepareConverter()
	if err != nil {
		return fmt.Errorf("failed to prepare converter: %v", err)
	}
	// Channel to manage the files to process
	fileChan := make(chan string)

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < parallelism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range fileChan {
				fmt.Printf("Processing file: %s\n", path)

				// Load the chapter
				chapter, err := cbz.LoadChapter(path)
				if err != nil {
					fmt.Printf("Failed to load chapter: %v\n", err)
					continue
				}

				if chapter.IsConverted {
					fmt.Printf("Chapter already converted: %s\n", path)
					continue
				}

				// Convert the chapter
				convertedChapter, err := chapterConverter.ConvertChapter(chapter, quality, func(msg string) {
					fmt.Println(msg)
				})
				if err != nil {
					fmt.Printf("Failed to convert chapter: %v\n", err)
					continue
				}
				convertedChapter.SetConverted()

				// Write the converted chapter back to a CBZ file
				outputPath := path
				if !override {
					outputPath = strings.TrimSuffix(path, ".cbz") + "_converted.cbz"
				}
				err = cbz.WriteChapterToCBZ(convertedChapter, outputPath)
				if err != nil {
					fmt.Printf("Failed to write converted chapter: %v\n", err)
					continue
				}

				fmt.Printf("Converted file written to: %s\n", outputPath)
			}
		}()
	}

	// Walk the path and send files to the channel
	err = filepath.WalkDir(path, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".cbz") {
			fileChan <- path
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking the path: %w", err)
	}

	close(fileChan) // Close the channel to signal workers to stop
	wg.Wait()       // Wait for all workers to finish

	return nil
}
