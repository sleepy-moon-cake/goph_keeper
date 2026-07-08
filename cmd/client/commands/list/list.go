package list

import (
	"context"
	"fmt"
	"goph_keeper/internal/client/interfaces"
	"goph_keeper/internal/shared/config"
	"goph_keeper/internal/shared/models"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var limit int

func NewListCmd(service interfaces.TransportService) *cobra.Command {
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "get entities, number is stricted by limit, default limit = 100",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			token, err := config.LoadToken()

			if err != nil {
				return fmt.Errorf("list command: %w", err)
			}

			if token == "" {
				return fmt.Errorf("you are not logged in. Please run 'gophkeeper login' first")
			}

			ctx = context.WithValue(ctx, models.TokenContextKey, token)

			records, err := service.ListRecords(ctx, limit)

			if err != nil {
				return fmt.Errorf("list command:%w", err)
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"#", "Name", "Type"})

			for i, rec := range records {
				table.Append([]string{
					fmt.Sprintf("%d", i+1),
					rec.Name,
					rec.DataType,
				})
			}

			fmt.Println("\n📋 YOUR SAVED RECORDS:")
			table.Render()
			fmt.Println()

			return nil
		},
	}

	listCmd.Flags().IntVar(&limit, "limit", 100, "limit")

	return listCmd
}
