package delete

import (
	"fmt"
	"goph_keeper/internal/client/interfaces"

	"github.com/spf13/cobra"
)

var name string

func NewDeleteCmd(service interfaces.TransportService) *cobra.Command {

	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "delete entity by name or id",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

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
