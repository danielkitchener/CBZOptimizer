package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	utils2 "github.com/danielkitchener/CBZOptimizer/v2/internal/utils"
	"github.com/danielkitchener/CBZOptimizer/v2/pkg/converter"
	"github.com/danielkitchener/CBZOptimizer/v2/pkg/converter/constant"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
)

var converterType constant.ConversionFormat

func init() {
	command := &cobra.Command{
		Use:   "optimize [folder]",
		Short: "Optimize all CBZ/CBR files in a folder recursively",
		Long:  "Optimize all CBZ/CBR files in a folder recursively.\nIt will take all the different pages in the CBZ/CBR files and convert them to the given format.\nThe original CBZ/CBR files will be kept intact depending if you choose to override or not.",
		RunE:  ConvertCbzCommand,
		Args:  cobra.ExactArgs(1),
	}
	formatFlag := enumflag.New(&converterType, "format", constant.CommandValue, enumflag.EnumCaseInsensitive)
	_ = formatFlag.RegisterCompletion(command, "format", constant.HelpText)

	command.Flags().Uint8P("quality", "q", 85, "Quality for conversion (0-100)")
	command.Flags().BoolP("lossless", "p", false, "Use lossless conversion (overrides quality setting)")
	command.MarkFlagsOneRequired("quality", "lossless")
	command.MarkFlagsMutuallyExclusive("quality", "lossless")
	command.Flags().IntP("parallelism", "n", 2, "Number of chapters to convert in parallel")
	command.Flags().BoolP("override", "o", false, "Override the original CBZ/CBR files")
	command.Flags().BoolP("split", "s", false, "Split long pages into smaller chunks")
	command.Flags().DurationP("timeout", "t", 0, "Maximum time allowed for converting a single chapter (e.g., 30s, 5m, 1h). 0 means no timeout")
	command.PersistentFlags().VarP(
		formatFlag,
		"format", "f",
		fmt.Sprintf("Format to convert the images to: %s", constant.ListAll()))
	command.PersistentFlags().Lookup("format").NoOptDefVal = constant.DefaultConversion.String()

	AddCommand(command)
}

func ConvertCbzCommand(cmd *cobra.Command, args []string) error {
	log.Info().Str("command", "optimize").Msg("Starting optimize command")

	path := args[0]
	if path == "" {
		log.Error().Msg("Path argument is required but empty")
		return fmt.Errorf("path is required")
	}

	log.Debug().Str("input_path", path).Msg("Validating input path")
	if !utils2.IsValidFolder(path) {
		log.Error().Str("input_path", path).Msg("Path validation failed - not a valid folder")
		return fmt.Errorf("the path needs to be a folder")
	}
	log.Debug().Str("input_path", path).Msg("Input path validated successfully")

	log.Debug().Msg("Parsing command-line flags")

	quality, err := cmd.Flags().GetUint8("quality")
	lossless, err2 := cmd.Flags().GetBool("lossless")
	if err != nil || err2 != nil || (!lossless && (quality <= 0 || quality > 100)) {
		log.Error().Err(err).Uint8("quality", quality).Msg("Invalid quality value")
		return fmt.Errorf("invalid quality value")
	}
	log.Debug().Uint8("quality", quality).Msg("Quality parameter validated")

	override, err := cmd.Flags().GetBool("override")
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse override flag")
		return fmt.Errorf("invalid quality value")
	}
	log.Debug().Bool("override", override).Msg("Override parameter parsed")

	split, err := cmd.Flags().GetBool("split")
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse split flag")
		return fmt.Errorf("invalid split value")
	}
	log.Debug().Bool("split", split).Msg("Split parameter parsed")

	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse timeout flag")
		return fmt.Errorf("invalid timeout value")
	}
	log.Debug().Dur("timeout", timeout).Msg("Timeout parameter parsed")

	parallelism, err := cmd.Flags().GetInt("parallelism")
	if err != nil || parallelism < 1 {
		log.Error().Err(err).Int("parallelism", parallelism).Msg("Invalid parallelism value")
		return fmt.Errorf("invalid parallelism value")
	}
	log.Debug().Int("parallelism", parallelism).Msg("Parallelism parameter validated")

	log.Debug().Str("converter_format", converterType.String()).Msg("Initializing converter")
	chapterConverter, err := converter.Get(converterType)
	if err != nil {
		log.Error().Str("converter_format", converterType.String()).Err(err).Msg("Failed to get chapter converter")
		return fmt.Errorf("failed to get chapterConverter: %v", err)
	}
	log.Debug().Str("converter_format", converterType.String()).Msg("Converter initialized successfully")

	log.Debug().Msg("Preparing converter")
	err = chapterConverter.PrepareConverter()
	if err != nil {
		log.Error().Err(err).Msg("Failed to prepare converter")
		return fmt.Errorf("failed to prepare converter: %v", err)
	}
	log.Debug().Msg("Converter prepared successfully")

	// Channel to manage the files to process
	fileChan := make(chan string)
	// Channel to collect errors
	errorChan := make(chan error, parallelism)

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start worker goroutines
	log.Debug().Int("worker_count", parallelism).Msg("Starting worker goroutines")
	for i := 0; i < parallelism; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Debug().Int("worker_id", workerID).Msg("Worker started")
			for path := range fileChan {
				log.Debug().Int("worker_id", workerID).Str("file_path", path).Msg("Worker processing file")
				err := utils2.Optimize(&utils2.OptimizeOptions{
					ChapterConverter: chapterConverter,
					Path:             path,
					Quality:          quality,
					Lossless:         lossless,
					Override:         override,
					Split:            split,
					Timeout:          timeout,
				})
				if err != nil {
					log.Error().Int("worker_id", workerID).Str("file_path", path).Err(err).Msg("Worker encountered error")
					errorChan <- fmt.Errorf("error processing file %s: %w", path, err)
				} else {
					log.Debug().Int("worker_id", workerID).Str("file_path", path).Msg("Worker completed file successfully")
				}
			}
			log.Debug().Int("worker_id", workerID).Msg("Worker finished")
		}(i)
	}
	log.Debug().Int("worker_count", parallelism).Msg("All worker goroutines started")

	// Walk the path and send files to the channel
	log.Debug().Str("search_path", path).Msg("Starting filesystem walk for CBZ/CBR files")
	err = filepath.WalkDir(path, func(filePath string, info os.DirEntry, err error) error {
		if err != nil {
			log.Error().Str("file_path", filePath).Err(err).Msg("Error during filesystem walk")
			return err
		}

		if !info.IsDir() {
			fileName := strings.ToLower(info.Name())
			if strings.HasSuffix(fileName, ".cbz") || strings.HasSuffix(fileName, ".cbr") {
				log.Debug().Str("file_path", filePath).Str("file_name", fileName).Msg("Found CBZ/CBR file")
				fileChan <- filePath
			}
		}

		return nil
	})

	if err != nil {
		log.Error().Str("search_path", path).Err(err).Msg("Filesystem walk failed")
		return fmt.Errorf("error walking the path: %w", err)
	}
	log.Debug().Str("search_path", path).Msg("Filesystem walk completed")

	close(fileChan) // Close the channel to signal workers to stop
	log.Debug().Msg("File channel closed, waiting for workers to complete")
	wg.Wait() // Wait for all workers to finish
	log.Debug().Msg("All workers completed")
	close(errorChan) // Close the error channel

	var errs []error
	for err := range errorChan {
		errs = append(errs, err)
		log.Error().Err(err).Msg("Collected processing error")
	}

	if len(errs) > 0 {
		log.Error().Int("error_count", len(errs)).Msg("Command completed with errors")
		return fmt.Errorf("encountered errors: %v", errs)
	}

	log.Info().Str("search_path", path).Msg("Optimize command completed successfully")
	return nil
}
