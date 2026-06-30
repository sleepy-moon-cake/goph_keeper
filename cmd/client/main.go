package main

import (
	"context"
	"fmt"
	"goph_keeper/cmd/client/commands"
	"log/slog"
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Start appication")

	if err := commands.Execute(ctx); err != nil {
		cancel()
		slog.Error("fatal error")
		fmt.Printf("err: %s", err)
		os.Exit(1)
	}

	fmt.Println("End appication")
}
