package delete

import (
	"context"
	"fmt"
	"goph_keeper/internal/client/interfaces"
	"goph_keeper/internal/shared/config"
	"goph_keeper/internal/shared/models"

	"github.com/spf13/cobra"
)

var name string

func NewDeleteCmd(service interfaces.TransportService) *cobra.Command {

	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "delete entity by name",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			token, err := config.LoadToken()

			if err != nil {
				return fmt.Errorf("delete command: %w", err)
			}

			if token == "" {
				return fmt.Errorf("you are not logged in. Please run 'gophkeeper login' first")
			}

			ctx = context.WithValue(ctx, models.TokenContextKey, token)

			if name == "" {
				return fmt.Errorf("name is required param")
			}

			if err := service.DeleteEntityByName(ctx, name); err != nil {
				return fmt.Errorf("delete entity:%w", err)
			}

			return nil
		},
	}

	deleteCmd.Flags().StringVar(&name, "name", "", "pass name")

	return deleteCmd
}
