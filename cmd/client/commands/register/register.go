package register

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

func NewRegisterCmd(service interfaces.TransportService) *cobra.Command {
	var registerCmd = &cobra.Command{
		Use:   "register",
		Short: "registration",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			fmt.Println("Registration:::::")

			fmt.Print("Enter username: ")
			scanner := bufio.NewScanner(os.Stdin)
			var userName string
			if scanner.Scan() {
				userName = strings.TrimSpace(scanner.Text())
			}
			if len(userName) == 0 {
				return fmt.Errorf("username cannot be empty")
			}

			fmt.Print("Enter password: ")
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			password := string(bytePassword)

			if len(password) == 0 {
				return fmt.Errorf("password cannot be empty")
			}

			fmt.Print("Enter confirm password: ")
			byteConfirmPassword, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("failed to read confirmation: %w", err)
			}
			confirmPassword := string(byteConfirmPassword)

			if password != confirmPassword {
				return fmt.Errorf("passwords are not equal")
			}

			h := sha256.New()
			h.Write([]byte(password))
			passwordHash := fmt.Sprintf("%x", h.Sum(nil))

			token, err := service.Register(ctx, userName, passwordHash)
			if err != nil {
				return fmt.Errorf("login command: %w", err)
			}

			if err := config.SaveToken(token, userName); err != nil {
				return fmt.Errorf("failed to save session: %w", err)
			}

			fmt.Println("Successfully registered!")
			return nil
		},
	}

	return registerCmd
}
