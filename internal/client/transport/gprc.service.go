package transport

import (
	"context"
	"fmt"
	"goph_keeper/internal/shared/models"
	"goph_keeper/internal/shared/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type GPRCTransportService struct {
	addr string
}

func NewGRPCTransportService(addr string) *GPRCTransportService {
	return &GPRCTransportService{
		addr: addr,
	}
}

func (t *GPRCTransportService) Register(ctx context.Context, name string, password string) (string, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	payload := &pb.AuthRequest{}
	payload.SetUsername(name)
	payload.SetPasswordHash(password)

	resp, err := client.Register(ctx, payload)
	if err != nil {
		return "", fmt.Errorf("login grpc: %w", err)
	}

	return resp.GetToken(), nil
}

func (t *GPRCTransportService) Login(ctx context.Context, name string, password string) (string, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	payload := &pb.AuthRequest{}
	payload.SetUsername(name)
	payload.SetPasswordHash(password)

	resp, err := client.Login(ctx, payload)
	if err != nil {
		return "", fmt.Errorf("login grpc: %w", err)
	}

	return resp.GetToken(), nil
}

func (t *GPRCTransportService) ListRecords(ctx context.Context, limit int) ([]models.RecordMeta, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	payload := &pb.ListRecordsRequest{}
	payload.SetLimit(int32(limit))

	resp, err := client.ListRecords(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("list records grpc: %w", err)
	}

	var result []models.RecordMeta
	for _, r := range resp.GetRecords() {
		result = append(result, models.RecordMeta{
			Name:     r.GetName(),
			DataType: r.GetDataType(),
		})
	}
	return result, nil
}

func (t *GPRCTransportService) GetEntityByName(ctx context.Context, name string) (*models.EncryptedRecord, error) {
	client, err := t.getClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	payload := &pb.GetRecordRequest{}
	payload.SetName(name)

	resp, err := client.GetRecord(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("get record: %w", err)
	}

	return &models.EncryptedRecord{
		Name:     resp.GetName(),
		DataType: resp.GetDataType(),
		Payload:  resp.GetSecureData(),
		Nonce:    resp.GetNonce(),
	}, nil
}

func (t *GPRCTransportService) SaveText(ctx context.Context, data models.TextData) error {
	client, err := t.getClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	encryptedBytes, nonce, err := getEncryptedData(data)
	if err != nil {
		return fmt.Errorf("saveText crypto: %w", err)
	}

	payload := &pb.Record{}
	payload.SetName(data.Name)
	payload.SetDataType("text")
	payload.SetSecureData(encryptedBytes)
	payload.SetNonce(nonce)

	if _, err := client.SaveRecord(ctx, payload); err != nil {
		return fmt.Errorf("saveText: %w", err)
	}
	return nil
}

func (t *GPRCTransportService) SaveCard(ctx context.Context, data models.CardData) error {
	client, err := t.getClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	encryptedBytes, nonce, err := getEncryptedData(data)
	if err != nil {
		return fmt.Errorf("saveCard crypto: %w", err)
	}

	payload := &pb.Record{}
	payload.SetName(data.Name)
	payload.SetDataType("card")
	payload.SetSecureData(encryptedBytes)
	payload.SetNonce(nonce)

	if _, err := client.SaveRecord(ctx, payload); err != nil {
		return fmt.Errorf("saveCard: %w", err)
	}
	return nil
}

func (t *GPRCTransportService) SaveFile(ctx context.Context, data models.BinaryData) error {
	client, err := t.getClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	encryptedBytes, nonce, err := getEncryptedData(data)
	if err != nil {
		return fmt.Errorf("saveFile crypto: %w", err)
	}

	payload := &pb.Record{}
	payload.SetName(data.Name)
	payload.SetDataType("file")
	payload.SetSecureData(encryptedBytes)
	payload.SetNonce(nonce)

	if _, err := client.SaveRecord(ctx, payload); err != nil {
		return fmt.Errorf("saveFile: %w", err)
	}
	return nil
}

func (t *GPRCTransportService) DeleteEntityByName(ctx context.Context, name string) error {
	client, err := t.getClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	payload := &pb.GetRecordRequest{}
	payload.SetName(name)

	if _, err := client.DeleteRecord(ctx, payload); err != nil {
		return fmt.Errorf("delete record: %w", err)
	}
	return nil
}

func (t *GPRCTransportService) getClient(ctx context.Context) (struct {
	pb.TransportServiceClient
	Close func() error
}, error) {

	conn, err := grpc.NewClient(t.addr, grpc.WithChainUnaryInterceptor(newAuthInterceptor(ctx)), grpc.WithTransportCredentials(insecure.NewCredentials()))

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

func newAuthInterceptor(cobraCtx context.Context) func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if token, ok := cobraCtx.Value(models.TokenContextKey).(string); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
