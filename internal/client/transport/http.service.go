package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"goph_keeper/internal/client/interfaces"
	"goph_keeper/internal/shared/models"
)

type HttpTransportService struct {
	addr       string
	httpClient *http.Client
	cache      interfaces.CacheService
	encryptKey string
}

type RoundTripperWrapper struct {
	http.RoundTripper
}

func (r *RoundTripperWrapper) RoundTrip(w *http.Request) (*http.Response, error) {
	if token, ok := models.GetToken(w.Context()); ok {
		w = w.WithContext(w.Context())
		w.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	return r.RoundTripper.RoundTrip(w)
}

func NewHttpTransportService(addr string, cache interfaces.CacheService, encryptKey string) *HttpTransportService {

	return &HttpTransportService{
		addr: addr,
		httpClient: &http.Client{
			Transport: &RoundTripperWrapper{
				RoundTripper: http.DefaultTransport,
			},
		},
		cache:      cache,
		encryptKey: encryptKey,
	}
}

func (t *HttpTransportService) Register(ctx context.Context, name string, password string) (string, error) {
	body := models.AuthRequest{Name: name, PasswordHash: password}

	var buffer bytes.Buffer

	if err := json.NewEncoder(&buffer).Encode(body); err != nil {
		return "", fmt.Errorf("failed to encode request body: %w", err)
	}

	registerURL := fmt.Sprintf("%s/api/v1/auth/register", t.addr)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, registerURL, &buffer)
	if err != nil {
		return "", fmt.Errorf("failed to form request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)

	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var session models.AuthResponse

	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		slog.Error("Login error: parsing body")
		return "", fmt.Errorf("cant parse body: %w", err)
	}

	return session.Session, nil
}

func (t *HttpTransportService) Login(ctx context.Context, name string, password string) (string, error) {
	body := models.AuthRequest{Name: name, PasswordHash: password}

	var buffer bytes.Buffer

	if err := json.NewEncoder(&buffer).Encode(body); err != nil {
		return "", fmt.Errorf("failed to encode: %w", err)
	}

	loginURL := fmt.Sprintf("%s/api/v1/auth/login", t.addr)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, &buffer)

	if err != nil {
		return "", fmt.Errorf("failed to form request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var session models.AuthResponse

	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		slog.Error("Login error: parsing body")
		return "", fmt.Errorf("cant parse body: %w", err)
	}

	return session.Session, nil
}

func (t *HttpTransportService) ListRecords(ctx context.Context, limit int) ([]models.RecordMeta, error) {
	fullURL := fmt.Sprintf("%s/api/v1/records?limit=%d", t.addr, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list request: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		slog.Warn("HTTP request failed, switching to local cache", "error", err)
		return t.cache.GetRecords(ctx)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var records []models.RecordMeta
	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		return nil, fmt.Errorf("failed to decode list response: %w", err)
	}

	if err := t.cache.UpdateRecords(ctx, records); err != nil {
		slog.Warn("failed to update local cache", "error", err)
	}

	return records, nil
}

func (t *HttpTransportService) GetEntityByName(ctx context.Context, name string) (*models.DecryptedRecord, error) {
	cryptedKey, ok := models.GetCryptedKey(ctx)

	if !ok {
		return nil, fmt.Errorf("get entity")
	}

	fullURL := fmt.Sprintf("%s/api/v1/records/%s", t.addr, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get request: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		slog.Warn("HTTP request failed, switching to local cache", "error", err)

		record, err := t.cache.GetRecordByName(ctx, name)

		if err != nil {
			return nil, fmt.Errorf("get record from cashe: %w", err)
		}

		return getDecryptedData(record, cryptedKey)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("record not found")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var record models.EncryptedRecord
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	if err := t.cache.UpdateSingleRecord(ctx, &record); err != nil {
		slog.Warn("failed to update local cache", "name", record.Name, "error", err)
	}

	return getDecryptedData(&record, cryptedKey)
}

func (t *HttpTransportService) DeleteEntityByName(ctx context.Context, name string) error {
	fullURL := fmt.Sprintf("%s/api/v1/records/%s", t.addr, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fullURL, nil)
	if err != nil {
		return fmt.Errorf("try to delete entity: %w", err)
	}

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

func (t *HttpTransportService) SaveText(ctx context.Context, data models.TextData) error {
	cryptedKey, ok := models.GetCryptedKey(ctx)
	if !ok {
		return fmt.Errorf("save text")
	}

	dst, nonce, err := getEncryptedData(data, cryptedKey)
	if err != nil {
		return fmt.Errorf("encrypt data: %w", err)
	}

	record := models.EncryptedRecord{
		Name:     data.Name,
		DataType: "text",
		Payload:  dst,
		Nonce:    nonce,
	}

	if err := sendJSON(ctx, t, record); err != nil {
		return fmt.Errorf("try to send data: %w", err)
	}
	return nil
}

func (t *HttpTransportService) SaveCard(ctx context.Context, data models.CardData) error {
	cryptedKey, ok := models.GetCryptedKey(ctx)
	if !ok {
		return fmt.Errorf("save card")
	}

	dst, nonce, err := getEncryptedData(data, cryptedKey)
	if err != nil {
		return fmt.Errorf("encrypt data: %w", err)
	}

	record := models.EncryptedRecord{
		Name:     data.Name,
		DataType: "card",
		Payload:  dst,
		Nonce:    nonce,
	}

	if err := sendJSON(ctx, t, record); err != nil {
		return fmt.Errorf("try to send data: %w", err)
	}
	return nil
}

func (t *HttpTransportService) SaveFile(ctx context.Context, data models.BinaryData) error {
	cryptedKey, ok := models.GetCryptedKey(ctx)
	if !ok {
		return fmt.Errorf("save file")
	}

	dst, nonce, err := getEncryptedData(data, cryptedKey)
	if err != nil {
		return fmt.Errorf("encrypt data: %w", err)
	}

	record := models.EncryptedRecord{
		Name:     data.Name,
		DataType: "file",
		Payload:  dst,
		Nonce:    nonce,
	}

	if err := sendJSON(ctx, t, record); err != nil {
		return fmt.Errorf("try to store file: %w", err)
	}
	return nil
}

func sendJSON(ctx context.Context, t *HttpTransportService, data models.EncryptedRecord) error {
	var bdata bytes.Buffer
	if err := json.NewEncoder(&bdata).Encode(data); err != nil {
		slog.Error("json encode error", "error", err)
		return fmt.Errorf("failed to encode data: %w", err)
	}

	fullURL := fmt.Sprintf("%s/api/v1/records", t.addr)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, &bdata)
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
