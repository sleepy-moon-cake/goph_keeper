package start

import (
	"fmt"
	"goph_keeper/cmd/client/session"
	"log/slog"

	"github.com/spf13/cobra"
)

func NewStartServerCmd(sm *session.SessionManager) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "launch session keeper",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("Запуск сервера сессий...")

			if err := sm.Listen(cmd.Context()); err != nil {
				return fmt.Errorf("server error: %w", err)
			}

			return nil
		},
	}
}
