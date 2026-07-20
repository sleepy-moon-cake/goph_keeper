package commands

import (
	"context"
	"fmt"
	"goph_keeper/cmd/client/commands/add"
	"goph_keeper/cmd/client/commands/delete"
	"goph_keeper/cmd/client/commands/get"
	"goph_keeper/cmd/client/commands/list"
	"goph_keeper/cmd/client/commands/login"
	"goph_keeper/cmd/client/commands/register"
	"goph_keeper/cmd/client/commands/start"
	"goph_keeper/cmd/client/session"
	"goph_keeper/internal/client/cache"
	"goph_keeper/internal/client/db"
	"goph_keeper/internal/client/transport"
	"goph_keeper/internal/shared/models"
	"log/slog"

	"github.com/spf13/cobra"
)

var (
	grpcAddr string

	serverAddr string

	sessionAddr string
)

var rootCmd = &cobra.Command{
	Use:   "gophkeeper",
	Short: "Gophkeeper - store manager",
}

func Execute(ctx context.Context) error {

	conn, err := db.NewConnector(ctx)
	if err != nil {
		return fmt.Errorf("new connections to cash db: %w", err)
	}

	cache := cache.NewCacheService(conn)

	ts, err := transport.NewTransportService(&transport.TransportConfig{
		AddrGRPC: grpcAddr,
		AddrHTTP: serverAddr,
		Cache:    cache,
	})

	if err != nil {
		return fmt.Errorf("create transport service:%w", err)
	}

	sm := session.NewSessionManager(sessionAddr)
	cm := session.NewClientSession(sessionAddr)

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		name := cmd.Name()

		if name == "start" || name == "help" {
			return nil
		}

		if err := cm.Ping(ctx); err != nil {
			return err
		}

		if name == "login" || name == "register" {
			return nil
		}

		session, err := cm.Get()

		if err != nil {
			slog.Error(err.Error())
			return fmt.Errorf("you are not logged in. Please run 'gophkeeper login' first")
		}

		if session.Token == "" {
			return fmt.Errorf("you are not logged in. Please run 'gophkeeper login' first")
		}

		ctx = models.WithToken(ctx, session.Token)
		ctx = models.WithUserName(ctx, session.UserName)
		ctx = models.WithCryptedKey(ctx, session.CryptedKey)
		cmd.SetContext(ctx)

		return nil
	}

	rootCmd.AddCommand(
		start.NewStartServerCmd(sm),
		register.NewRegisterCmd(ts, func(name, key, token string) error {
			return cm.Set(name, key, token)
		}),
		login.NewLoginCmd(ts, func(name, key, token string) error {
			return cm.Set(name, key, token)
		}),
		add.NewAddCommand(ts),
		delete.NewDeleteCmd(ts),
		get.NewGetCmd(ts),
		list.NewListCmd(ts),
	)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return fmt.Errorf("execute cobra:%w", err)
	}
	return nil
}

func init() {
	// setup configuration flags
	rootCmd.PersistentFlags().StringVarP(&serverAddr, "http", "a", ":8080", "Http server address")
	rootCmd.PersistentFlags().StringVarP(&grpcAddr, "grpc", "g", ":3200", "gRPC server address")
	rootCmd.PersistentFlags().StringVarP(&sessionAddr, "session", "s", "127.0.0.1:8088", "session server address")
}
