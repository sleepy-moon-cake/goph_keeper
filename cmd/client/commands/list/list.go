package list

import (
	"fmt"
	"goph_keeper/internal/client/interfaces"

	"github.com/spf13/cobra"
)

var limit int

func NewListCmd(service interfaces.TransportService) *cobra.Command {
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "get entities, number is stricted by limit, default limit = 100",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if _, err := service.ListRecords(ctx, limit); err != nil {
				return fmt.Errorf("list command:%w", err)
			}

			return nil
		},
	}

	listCmd.Flags().IntVar(&limit, "limit", 100, "limit")

	return listCmd
}
