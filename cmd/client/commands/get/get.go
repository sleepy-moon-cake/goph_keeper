package get

import (
	"fmt"
	"goph_keeper/internal/client/interfaces"
	"goph_keeper/internal/shared/models"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var name string

func NewGetCmd(service interfaces.TransportService) *cobra.Command {
	var getCmd = &cobra.Command{
		Use:   "get",
		Short: "get entity by name",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if name == "" {
				return fmt.Errorf("name is required param")
			}

			entity, err := service.GetEntityByName(ctx, name)
			if err != nil {
				return fmt.Errorf("get entity: %w", err)
			}

			var dataString string
			switch entity.DataType {
			case "text":
				if textData, ok := entity.Data.(models.TextData); ok {
					dataString = textData.Text
				} else {
					dataString = "ошибка приведения типов текста"
				}

			case "card":
				if cardData, ok := entity.Data.(models.CardData); ok {
					dataString = fmt.Sprintf("Номер: %s | Срок: %s | CVV: %s", cardData.CardNumber, cardData.ExpirationDate, cardData.CVV)
				} else {
					dataString = "ошибка приведения типов карты"
				}

			case "file":
				if binaryData, ok := entity.Data.(models.BinaryData); ok {
					dataString = fmt.Sprintf("Файл: %s (%d байт)", binaryData.FileName, len(binaryData.Data))
				} else {
					dataString = "ошибка приведения типов файла"
				}

			default:
				dataString = "неизвестный формат данных"
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"#", "Name", "Type", "Data"})

			table.Append([]string{
				"1",
				entity.Name,
				entity.DataType,
				dataString,
			})

			fmt.Println("\n📋 YOUR SAVED RECORDS:")
			table.Render()
			fmt.Println()

			return nil
		},
	}

	getCmd.Flags().StringVar(&name, "name", "", "pass name")

	return getCmd
}
