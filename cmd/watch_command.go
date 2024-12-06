package cmd

import (
	"fmt"
	"github.com/belphemur/CBZOptimizer/v2/converter"
	"github.com/belphemur/CBZOptimizer/v2/converter/constant"
	"github.com/belphemur/CBZOptimizer/v2/utils"
	"github.com/pablodz/inotifywaitgo/inotifywaitgo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thediveo/enumflag/v2"
	"log"
	"runtime"
	"strings"
	"sync"
)

func init() {
	if runtime.GOOS != "linux" {
		return
	}
	command := &cobra.Command{
		Use:   "watch [folder]",
		Short: "Watch a folder for new CBZ files",
		Long:  "Watch a folder for new CBZ files.\nIt will watch a folder for new CBZ files and optimize them.",
		RunE:  WatchCommand,
		Args:  cobra.ExactArgs(1),
	}
	formatFlag := enumflag.New(&converterType, "format", constant.CommandValue, enumflag.EnumCaseInsensitive)
	_ = formatFlag.RegisterCompletion(command, "format", constant.HelpText)

	command.Flags().Uint8P("quality", "q", 85, "Quality for conversion (0-100)")
	_ = viper.BindPFlag("quality", command.Flags().Lookup("quality"))

	command.Flags().BoolP("override", "o", true, "Override the original CBZ files")
	_ = viper.BindPFlag("override", command.Flags().Lookup("override"))

	command.Flags().BoolP("split", "s", false, "Split long pages into smaller chunks")
	_ = viper.BindPFlag("split", command.Flags().Lookup("split"))

	command.PersistentFlags().VarP(
		formatFlag,
		"format", "f",
		fmt.Sprintf("Format to convert the images to: %s", constant.ListAll()))
	command.PersistentFlags().Lookup("format").NoOptDefVal = constant.DefaultConversion.String()
	_ = viper.BindPFlag("format", command.PersistentFlags().Lookup("format"))

	AddCommand(command)
}
func WatchCommand(_ *cobra.Command, args []string) error {
	path := args[0]
	if path == "" {
		return fmt.Errorf("path is required")
	}

	if !utils.IsValidFolder(path) {
		return fmt.Errorf("the path needs to be a folder")
	}

	quality := uint8(viper.GetUint16("quality"))
	if quality <= 0 || quality > 100 {
		return fmt.Errorf("invalid quality value")
	}

	override := viper.GetBool("override")

	split := viper.GetBool("split")

	converterType := constant.FindConversionFormat(viper.GetString("format"))
	chapterConverter, err := converter.Get(converterType)
	if err != nil {
		return fmt.Errorf("failed to get chapterConverter: %v", err)
	}

	err = chapterConverter.PrepareConverter()
	if err != nil {
		return fmt.Errorf("failed to prepare converter: %v", err)
	}
	log.Printf("Watching [%s] with [override: %t, quality: %d, format: %s, split: %t]", path, override, quality, converterType.String(), split)

	events := make(chan inotifywaitgo.FileEvent)
	errors := make(chan error)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		inotifywaitgo.WatchPath(&inotifywaitgo.Settings{
			Dir:        path,
			FileEvents: events,
			ErrorChan:  errors,
			Options: &inotifywaitgo.Options{
				Recursive: true,
				Events: []inotifywaitgo.EVENT{
					inotifywaitgo.MOVE,
					inotifywaitgo.CLOSE_WRITE,
				},
				Monitor: true,
			},
			Verbose: true,
		})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for event := range events {
			log.Printf("[Event]%s, %v\n", event.Filename, event.Events)

			if !strings.HasSuffix(strings.ToLower(event.Filename), ".cbz") {
				continue
			}

			for _, e := range event.Events {
				switch e {
				case inotifywaitgo.CLOSE_WRITE, inotifywaitgo.MOVE:
					err := utils.Optimize(&utils.OptimizeOptions{
						ChapterConverter: chapterConverter,
						Path:             event.Filename,
						Quality:          quality,
						Override:         override,
						Split:            split,
					})
					if err != nil {
						errors <- fmt.Errorf("error processing file %s: %w", event.Filename, err)
					}
				default:
					// ignored
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range errors {
			log.Printf("Error: %v\n", err)
		}
	}()

	wg.Wait()
	return nil
}
