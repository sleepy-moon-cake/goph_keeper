package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"goph_keeper/internal/shared/models"
	"log/slog"
	"net/http"
	"strings"
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

func (t *HttpTransportService) DeleteEntityByName(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, t.addr, strings.NewReader(name))

	if err != nil {
		return fmt.Errorf("try to delete entity: %w", err)
	}
	req.Header.Set("Content-type", "text/plain")

	resp, err := t.httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	return nil
}

func (t *HttpTransportService) SaveText(ctx context.Context, data models.TextData) error {
	if err := sendJSON(ctx, t, data); err != nil {
		return fmt.Errorf("try to store text: %w", err)
	}
	return nil
}

func (t *HttpTransportService) SaveCard(ctx context.Context, data models.CardData) error {
	if err := sendJSON(ctx, t, data); err != nil {
		return fmt.Errorf("try to store card: %w", err)
	}
	return nil
}

func (t *HttpTransportService) SaveFile(ctx context.Context, data models.BinaryData) error {
	if err := sendJSON(ctx, t, data); err != nil {
		return fmt.Errorf("try to store file: %w", err)
	}
	return nil
}

func sendJSON[T models.BinaryData | models.CardData | models.TextData](ctx context.Context, t *HttpTransportService, data T) error {
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
