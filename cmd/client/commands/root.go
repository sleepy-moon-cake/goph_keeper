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
	"goph_keeper/internal/client/cache"
	"goph_keeper/internal/client/db"
	"goph_keeper/internal/client/transport"

	"github.com/spf13/cobra"
)

var (
	grpcAddr string

	serverAddr string
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

	rootCmd.AddCommand(
		register.NewRegisterCmd(ts),
		login.NewLoginCmd(ts),
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
}
