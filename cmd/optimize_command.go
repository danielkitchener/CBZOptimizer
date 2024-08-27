package cmd

import (
	"fmt"
	"github.com/belphemur/CBZOptimizer/converter"
	"github.com/belphemur/CBZOptimizer/converter/constant"
	"github.com/belphemur/CBZOptimizer/utils"
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

func ConvertCbzCommand(cmd *cobra.Command, args []string) error {
	path := args[0]
	if path == "" {
		return fmt.Errorf("path is required")
	}

	if !utils.IsValidFolder(path) {
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
	// Channel to collect errors
	errorChan := make(chan error, parallelism)

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < parallelism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range fileChan {
				err := utils.Optimize(chapterConverter, path, quality, override)
				if err != nil {
					errorChan <- fmt.Errorf("error processing file %s: %w", path, err)
				}
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

	close(fileChan)  // Close the channel to signal workers to stop
	wg.Wait()        // Wait for all workers to finish
	close(errorChan) // Close the error channel

	var errs []error
	for err := range errorChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("encountered errors: %v", errs)
	}

	return nil
}
