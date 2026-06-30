package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"goph_keeper/internal/shared/models"
	"log/slog"
	"net/http"
)

type HttpTransportService struct {
	addr       string
	httpClient *http.Client
}

func NewHttpTransportService(addr string) *HttpTransportService {
	return &HttpTransportService{
		addr:       addr,
		httpClient: &http.Client{},
	}
}

func (t *HttpTransportService) SaveText(ctx context.Context, data models.TextData) error {
	if err := t.sendJSON(ctx, data); err != nil {
		return fmt.Errorf("try to store text: %w", err)
	}
	return nil
}

func (t *HttpTransportService) SaveCard(ctx context.Context, data models.CardData) error {
	if err := t.sendJSON(ctx, data); err != nil {
		return fmt.Errorf("try to store card: %w", err)
	}
	return nil
}

func (t *HttpTransportService) SaveFile(ctx context.Context, data models.BinaryData) error {
	if err := t.sendJSON(ctx, data); err != nil {
		return fmt.Errorf("try to store file: %w", err)
	}
	return nil
}

func (t *HttpTransportService) sendJSON(ctx context.Context, data interface{}) error {
	var bdata bytes.Buffer
	if err := json.NewEncoder(&bdata).Encode(data); err != nil {
		slog.Error("json encode error", "error", err)
		return fmt.Errorf("failed to encode data: %w", err)
	} 

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.addr, &bdata)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	return nil
}
