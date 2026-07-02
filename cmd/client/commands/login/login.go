package login

import (
	"crypto/sha256"
	"fmt"
	"goph_keeper/internal/client/interfaces"
	"goph_keeper/internal/shared/config"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var name string

func NewLoginCmd(service interfaces.TransportService) {

	var loginCmd = &cobra.Command{
		Use:   "login",
		Short: "authorization",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if name == "" {
				return fmt.Errorf("--username is required")
			}

			fmt.Print("Enter password: ")
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Println()

			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}

			if len(bytePassword) == 0 {
				return fmt.Errorf("password cannot be empty")
			}

			h := sha256.New()
			h.Write(bytePassword)
			passwordHash := fmt.Sprintf("%x", h.Sum(nil))

			token, err := service.Login(ctx, name, passwordHash)
			if err != nil {
				return fmt.Errorf("login command:%w", err)
			}

			if err := config.SaveToken(token); err != nil {
				return fmt.Errorf("failed to save session: %w", err)
			}

			fmt.Println("Successfully logged in!")
			return nil
		},
	}

	loginCmd.Flags().StringVar(&name, "name", "", "pass name")
}
