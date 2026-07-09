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

			session, err := config.LoadSession()

			if err != nil {
				return fmt.Errorf("add command: %w", err)
			}

			if session.Token == "" {
				return fmt.Errorf("you are not logged in. Please run 'gophkeeper login' first")
			}

			ctx = context.WithValue(ctx, models.TokenContextKey, session.Token)
			ctx = context.WithValue(ctx, models.UserContextKey, session.UserName)

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
