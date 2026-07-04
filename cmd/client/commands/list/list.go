package list

import (
	"context"
	"fmt"
	"goph_keeper/internal/client/interfaces"
	"goph_keeper/internal/shared/config"
	"goph_keeper/internal/shared/models"

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

			if _, err := service.ListRecords(ctx, limit); err != nil {
				return fmt.Errorf("list command:%w", err)
			}

			return nil
		},
	}

	listCmd.Flags().IntVar(&limit, "limit", 100, "limit")

	return listCmd
}
