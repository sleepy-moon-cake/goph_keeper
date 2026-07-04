package get

import (
	"context"
	"fmt"
	"goph_keeper/internal/client/interfaces"
	"goph_keeper/internal/shared/config"
	"goph_keeper/internal/shared/models"

	"github.com/spf13/cobra"
)

var name string

func NewGetCmd(service interfaces.TransportService) *cobra.Command {
	var getCmd = &cobra.Command{
		Use:   "get",
		Short: "get entity by name",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			token, err := config.LoadToken()

			if err != nil {
				return fmt.Errorf("get command: %w", err)
			}

			if token == "" {
				return fmt.Errorf("you are not logged in. Please run 'gophkeeper login' first")
			}

			ctx = context.WithValue(ctx, models.TokenContextKey, token)

			if name == "" {
				return fmt.Errorf("name is required param")
			}

			if _, err := service.GetEntityByName(ctx, name); err != nil {
				return fmt.Errorf("delete entity:%w", err)
			}

			return nil
		},
	}

	getCmd.Flags().StringVar(&name, "name", "", "pass name")

	return getCmd
}
