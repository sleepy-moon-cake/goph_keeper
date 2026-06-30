package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	isText bool
	isFile bool
	isCard bool
)

var (
	name string
	path string
)

var maxMemorySize = 10 * 1024 * 1024

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "store",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if !isCard && !isFile && !isText {
			return fmt.Errorf("assign data type (--text, --card or --file)")
		}

		if name == "" {
			return fmt.Errorf("name is required param")
		}

		switch {
		case isText:
			text, err := handleText()
			if err != nil {
				return err
			}

			if err := clientService.SaveText(ctx, text); err != nil {
				return fmt.Errorf("saveText %w", err)
			}

		case isFile:
			file, err := handleFile()
			if err != nil {
				return err
			}

			if err := clientService.SaveFile(ctx, file); err != nil {
				return fmt.Errorf("saveFile %w", err)
			}

		case isCard:
			card, err := handleCard()
			if err != nil {
				return err
			}

			if err := clientService.SaveCard(ctx, card); err != nil {
				return fmt.Errorf("saveCard %w", err)
			}
		}
		return nil
	},
}

func init() {
	// supported data
	addCmd.Flags().BoolVar(&isText, "text", false, "To store data in text format")
	addCmd.Flags().BoolVar(&isFile, "file", false, "To store binary file like pdf,doc")
	addCmd.Flags().BoolVar(&isCard, "card", false, "To store credit card params")

	// common flags
	addCmd.Flags().StringVar(&name, "name", "", "Set unique key")
	addCmd.Flags().StringVar(&path, "path", "", "Set path to file")

	rootCmd.AddCommand(addCmd)
}
