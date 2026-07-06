package main

import (
	"context"
	"fmt"
	"goph_keeper/internal/server/delivery"
	"goph_keeper/internal/server/middlewares"
	"goph_keeper/internal/shared/pb"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	secretKey := "super_secret_goph_keeper_key"
	grpcAddr := ":4200"
	httpAddr := ":8080"
	db := NewMockMemoryStorage()

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		fmt.Println("Start grpc server::::")
		grpcHandler := delivery.NewGRPCHandler(db, secretKey)

		if err := runGRPC(gCtx, grpcAddr, grpcHandler, secretKey); err != nil {
			return fmt.Errorf("gRPC server error: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		httpRouter := delivery.NewRouter(delivery.NewHTTPHandler(db, secretKey), secretKey)
		if err := runHttp(gCtx, httpAddr, httpRouter); err != nil {
			return fmt.Errorf("http server error: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		slog.Error("App error", "err", err)
	} else {
		slog.Info("Backend successfully stopped")
	}
}

func runHttp(ctx context.Context, addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,   // время на чтение запроса
		WriteTimeout: 10 * time.Second,  // время на отправку ответа
		IdleTimeout:  120 * time.Second, // время удержания соединения (Keep-Alive)
	}

	go func() {
		<-ctx.Done()

		ctxWithTime, cancel := context.WithTimeout(context.Background(), time.Second*5)

		defer cancel()

		srv.Shutdown(ctxWithTime)
	}()

	if err := srv.ListenAndServeTLS("cert.pem", "key.pem"); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		return fmt.Errorf("ListenAndServeTLS:%w", err)
	}

	return nil
}

func runGRPC(ctx context.Context, addr string, handler pb.TransportServiceServer, secretKey string) error {
	gctx, cancel := context.WithCancel(ctx)
	defer cancel()

	grpcListener, err := net.Listen("tcp", addr)

	if err != nil {
		return fmt.Errorf("create listener: %w", err)
	}

	gsrv := grpc.NewServer(
		grpc.UnaryInterceptor(middlewares.AuthUnaryInterceptor(secretKey)),
	)

	pb.RegisterTransportServiceServer(gsrv, handler)

	go func() {
		<-gctx.Done()

		gracefulStopDone := make(chan struct{})

		go func() {
			gsrv.GracefulStop()
			close(gracefulStopDone)
		}()

		ctxWithTime, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		select {
		case <-gracefulStopDone:
		case <-ctxWithTime.Done():
			gsrv.Stop()
		}
	}()

	if err := gsrv.Serve(grpcListener); err != nil && err != grpc.ErrServerStopped {
		return fmt.Errorf("grpcListener:%w", err)
	}

	return nil
}
