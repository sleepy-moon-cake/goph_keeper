package login

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"goph_keeper/internal/client/interfaces"
	"goph_keeper/internal/shared/config"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewLoginCmd(service interfaces.TransportService) *cobra.Command {
	var loginCmd = &cobra.Command{
		Use:   "login",
		Short: "authorization",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			fmt.Print("Enter username: ")
			scanner := bufio.NewScanner(os.Stdin)
			var userName string
			if scanner.Scan() {
				userName = strings.TrimSpace(scanner.Text())
			}

			if userName == "" {
				return fmt.Errorf("username cannot be empty")
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

			token, err := service.Login(ctx, userName, passwordHash)
			if err != nil {
				return fmt.Errorf("login command: %w", err)
			}

			if err := config.SaveToken(token, userName); err != nil {
				return fmt.Errorf("failed to save session: %w", err)
			}

			fmt.Println("Successfully logged in!")
			return nil
		},
	}

	return loginCmd
}
