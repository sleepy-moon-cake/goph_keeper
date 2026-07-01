package commands

import (
	"context"
	"fmt"
	"goph_keeper/cmd/client/commands/add"
	"goph_keeper/cmd/client/commands/delete"
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
	ts, err := transport.NewTransportService(&transport.TransportConfig{
		AddrGRPC: grpcAddr,
		AddrHTTP: serverAddr,
	})

	if err != nil {
		return fmt.Errorf("create trasport service:%w", err)
	}

	rootCmd.AddCommand(add.NewAddCommand(ts), delete.NewDeleteCmd(ts))

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return fmt.Errorf("execute cobra:%w", err)
	}
	return nil
}

func init() {
	// setup configuration flags
	rootCmd.PersistentFlags().StringVarP(&serverAddr, "http", "a", "http://localhost:4200", "Http server address")
	rootCmd.PersistentFlags().StringVarP(&grpcAddr, "grpc", "g", ":3200", "gRPC server address")
}
