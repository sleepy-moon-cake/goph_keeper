package transport

import (
	"context"
	"fmt"
	"goph_keeper/internal/shared/models"
	"goph_keeper/internal/shared/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GPRCTransportService struct {
	addr string
}

func NewGRPCTransportService(addr string) *GPRCTransportService {
	return &GPRCTransportService{
		addr: addr,
	}
}

func (t *GPRCTransportService) DeleteEntityByName(ctx context.Context, name string) error {
	client, err := t.getClient()

	if err != nil {
		return err
	}
	defer client.Close()

	payload := &pb.DeleteRequest{}
	payload.SetName(name)

	if _, err := client.DeleteEntityByName(ctx, payload); err != nil {
		return fmt.Errorf("delete entity:%w", err)
	}
	return nil
}

func (t *GPRCTransportService) SaveText(ctx context.Context, data models.TextData) error {
	client, err := t.getClient()

	if err != nil {
		return err
	}
	defer client.Close()

	payload := &pb.TextData{}

	payload.SetName(data.Name)
	payload.SetText(data.Text)

	if _, err := client.SaveText(ctx, payload); err != nil {
		return fmt.Errorf("saveText:%w", err)
	}
	return nil
}

func (t *GPRCTransportService) SaveCard(ctx context.Context, data models.CardData) error {
	client, err := t.getClient()

	if err != nil {
		return err
	}
	defer client.Close()

	payload := &pb.CardData{}

	payload.SetName(data.Name)
	payload.SetCardNumber(data.CardNumber)
	payload.SetCardholderName(data.CardholderName)
	payload.SetExpirationDate(data.ExpirationDate)
	payload.SetCvv(data.CVV)

	if _, err := client.SaveCard(ctx, payload); err != nil {
		return fmt.Errorf("saveCard: %w", err)
	}
	return nil
}

func (t *GPRCTransportService) SaveFile(ctx context.Context, data models.BinaryData) error {
	client, err := t.getClient()

	if err != nil {
		return err
	}
	defer client.Close()

	payload := &pb.BinaryData{}

	payload.SetName(data.Name)
	payload.SetFileName(data.FileName)
	payload.SetData(data.Data)

	if _, err := client.SaveFile(ctx, payload); err != nil {
		return fmt.Errorf("saveFile: %w", err)
	}
	return nil
}

func (t *GPRCTransportService) getClient() (struct {
	pb.TransportServiceClient
	Close func() error
}, error) {
	conn, err := grpc.NewClient(t.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return struct {
			pb.TransportServiceClient
			Close func() error
		}{}, fmt.Errorf("grpc: connection: %w", err)
	}

	return struct {
		pb.TransportServiceClient
		Close func() error
	}{
		TransportServiceClient: pb.NewTransportServiceClient(conn),
		Close:                  conn.Close,
	}, nil
}
