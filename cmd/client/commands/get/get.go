package get

import (
	"fmt"
	"goph_keeper/internal/client/interfaces"

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

			if _, err := service.GetEntityByName(ctx, name); err != nil {
				return fmt.Errorf("delete entity:%w", err)
			}

			return nil
		},
	}

	getCmd.Flags().StringVar(&name, "name", "", "pass name")

	return getCmd
}
