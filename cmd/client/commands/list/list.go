package list

import (
	"fmt"
	"goph_keeper/internal/client/interfaces"
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
