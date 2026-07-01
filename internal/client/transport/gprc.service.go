package transport

import (
	"context"
	"crypto/sha256"
	"encoding/json"
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

// Вспомогательная дженерик-функция для генерации зашифрованной (хэшированной) записи
func getEncryptedPayload[T models.CardData | models.TextData | models.BinaryData](data T) ([]byte, error) {
	rawBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	// Ваша временная хэш-заглушка вместо полноценного AES-GCM
	h := sha256.New()
	h.Write(rawBytes)
	dst := h.Sum(nil)

	return dst, nil
}

func (t *GPRCTransportService) GetEntityByName(ctx context.Context, name string) (*models.EncryptedRecord, error) {
	client, err := t.getClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// 1. Формируем запрос с именем сущности
	payload := &pb.GetRecordRequest{}
	payload.SetName(name)

	// 2. Запрашиваем "слепые" данные у сервера
	resp, err := client.GetRecord(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("get record: %w", err)
	}

	// 3. Собираем локальную модель зашифрованной записи через Opaque-геттеры
	return &models.EncryptedRecord{
		Name:     resp.GetName(),
		DataType: resp.GetDataType(),
		Payload:  resp.GetSecureData(),
		Nonce:    resp.GetNonce(),
	}, nil
}

func (t *GPRCTransportService) SaveText(ctx context.Context, data models.TextData) error {
	client, err := t.getClient()
	if err != nil {
		return err
	}
	defer client.Close()

	// 1. Хэшируем данные
	encryptedBytes, err := getEncryptedPayload(data)
	if err != nil {
		return fmt.Errorf("saveText crypto: %w", err)
	}

	// 2. Формируем универсальную gRPC-модель pb.Record
	payload := &pb.Record{}
	payload.SetName(data.Name)
	payload.SetDataType("text")
	payload.SetSecureData(encryptedBytes)
	payload.SetNonce(nil) // Для хэша nonce не нужен

	// 3. Вызываем единый gRPC метод сохранения
	if _, err := client.SaveRecord(ctx, payload); err != nil {
		return fmt.Errorf("saveText: %w", err)
	}
	return nil
}

func (t *GPRCTransportService) SaveCard(ctx context.Context, data models.CardData) error {
	client, err := t.getClient()
	if err != nil {
		return err
	}
	defer client.Close()

	encryptedBytes, err := getEncryptedPayload(data)
	if err != nil {
		return fmt.Errorf("saveCard crypto: %w", err)
	}

	payload := &pb.Record{}
	payload.SetName(data.Name)
	payload.SetDataType("card")
	payload.SetSecureData(encryptedBytes)
	payload.SetNonce(nil)

	if _, err := client.SaveRecord(ctx, payload); err != nil {
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

	encryptedBytes, err := getEncryptedPayload(data)
	if err != nil {
		return fmt.Errorf("saveFile crypto: %w", err)
	}

	payload := &pb.Record{}
	payload.SetName(data.Name)
	payload.SetDataType("file")
	payload.SetSecureData(encryptedBytes)
	payload.SetNonce(nil)

	if _, err := client.SaveRecord(ctx, payload); err != nil {
		return fmt.Errorf("saveFile: %w", err)
	}
	return nil
}

func (t *GPRCTransportService) DeleteEntityByName(ctx context.Context, name string) error {
	client, err := t.getClient()
	if err != nil {
		return err
	}
	defer client.Close()

	// Используем обновленный тип запроса pb.GetRecordRequest из нового прото-файла
	payload := &pb.GetRecordRequest{}
	payload.SetName(name)

	if _, err := client.DeleteRecord(ctx, payload); err != nil {
		return fmt.Errorf("delete record: %w", err)
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
