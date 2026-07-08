package add

import (
	"context"
	"fmt"
	"goph_keeper/internal/client/interfaces"
	"goph_keeper/internal/shared/config"
	"goph_keeper/internal/shared/models"

	"github.com/spf13/cobra"
)

var (
	isText bool
	isFile bool
	isCard bool
)

var (
	name   string
	path   string
	value  string
	holder string
)

var maxMemorySize = 10 * 1024 * 1024

// addCmd represents the add command

func NewAddCommand(service interfaces.TransportService) *cobra.Command {
	var addCmd = &cobra.Command{
		Use:   "add",
		Short: "store",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			token, err := config.LoadToken()

			if err != nil {
				return fmt.Errorf("add command: %w", err)
			}

			if token == "" {
				return fmt.Errorf("you are not logged in. Please run 'gophkeeper login' first")
			}

			ctx = context.WithValue(ctx, models.TokenContextKey, token)

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

				if err := service.SaveText(ctx, text); err != nil {
					return fmt.Errorf("saveText %w", err)
				}

			case isFile:
				file, err := handleFile()
				if err != nil {
					return err
				}

				if err := service.SaveFile(ctx, file); err != nil {
					return fmt.Errorf("saveFile %w", err)
				}

			case isCard:
				card, err := handleCard()
				if err != nil {
					return err
				}

				if err := service.SaveCard(ctx, card); err != nil {
					return fmt.Errorf("saveCard %w", err)
				}
			}
			return nil
		},
	}

	// supported data
	addCmd.Flags().BoolVar(&isText, "text", false, "To store data in text format")
	addCmd.Flags().BoolVar(&isFile, "file", false, "To store binary file like pdf,doc")
	addCmd.Flags().BoolVar(&isCard, "card", false, "To store credit card params")

	// common flags
	addCmd.Flags().StringVar(&name, "name", "", "Set unique key")
	addCmd.Flags().StringVar(&path, "path", "", "Set path to file")

	// text flags
	addCmd.Flags().StringVar(&value, "value", "", "Text content")

	// card flags
	addCmd.Flags().StringVar(&number, "number", "", "set credit card number")
	addCmd.Flags().StringVar(&cvv, "cvv", "", "set credit card cvv")
	addCmd.Flags().StringVar(&expire, "expire", "", "set credit card expired")
	addCmd.Flags().StringVar(&holder, "h", "card_holder", "set holder of card")

	return addCmd
}
