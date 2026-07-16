package delivery

import (
	"encoding/json"
	"goph_keeper/internal/server/interfaces"
	"goph_keeper/internal/server/middlewares"
	"goph_keeper/internal/server/utils"
	"goph_keeper/internal/shared/models"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type HTTPHandler struct {
	db        interfaces.RepositoryDb
	secretKey string
}

func NewRouter(handler *HTTPHandler, secretKey string) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/register", handler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", handler.Login)

	authWrapper := middlewares.JWTSession(secretKey)

	mux.Handle("GET /api/v1/records", authWrapper(http.HandlerFunc(handler.ListRecords)))
	mux.Handle("POST /api/v1/records", authWrapper(http.HandlerFunc(handler.SaveRecord)))

	mux.Handle("GET /api/v1/records/{name}", authWrapper(http.HandlerFunc(handler.GetRecord)))
	mux.Handle("DELETE /api/v1/records/{name}", authWrapper(http.HandlerFunc(handler.DeleteRecord)))

	return mux
}

func NewHTTPHandler(db interfaces.RepositoryDb, secretKey string) *HTTPHandler {
	return &HTTPHandler{
		db:        db,
		secretKey: secretKey,
	}
}

func (h *HTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.PasswordHash == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	if err := h.db.AddUser(r.Context(), req.Name, req.PasswordHash); err != nil {
		slog.Error("failed to register user", "error", err)

		http.Error(w, "Username is already taken", http.StatusConflict)
		return
	}

	token, err := utils.GenerateToken(req.Name, h.secretKey)
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(models.AuthResponse{Session: token}); err != nil {
		slog.Error("failed to encode token", "error", err)
	}
}

func (h *HTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	savedHash, err := h.db.GetUserPassword(r.Context(), req.Name)
	if err != nil || savedHash != req.PasswordHash {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateToken(req.Name, h.secretKey)
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(models.AuthResponse{Session: token}); err != nil {
		slog.Error("failed to encode token", "error", err)
	}
}

func (h *HTTPHandler) SaveRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username, ok := r.Context().Value(models.UserContextKey).(string)
	if !ok || username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var record models.EncryptedRecord
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if record.Name == "" {
		http.Error(w, "Record name is required", http.StatusBadRequest)
		return
	}

	if err := h.db.SaveRecord(r.Context(), username, record); err != nil {
		slog.Error("database error on save record", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HTTPHandler) GetRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username, ok := r.Context().Value(models.UserContextKey).(string)
	if !ok || username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/api/v1/records/")
	if name == "" {
		http.Error(w, "Record name is required", http.StatusBadRequest)
		return
	}

	dbRecord, err := h.db.GetRecord(r.Context(), username, name)
	if err != nil {
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dbRecord)
}

// 5. DELETE RECORD (DELETE /api/v1/records/{name})
func (h *HTTPHandler) DeleteRecord(w http.ResponseWriter, r *http.Request) {
	// (Реализуется зеркально методу GetRecord, проверяя http.MethodDelete)
}

// 6. LIST RECORDS (GET /api/v1/records?limit=10)
func (h *HTTPHandler) ListRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username, ok := r.Context().Value(models.UserContextKey).(string)
	if !ok || username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)

	dbMetas, err := h.db.ListRecords(r.Context(), username, int32(limit))
	if err != nil {
		slog.Error("database error on list records", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dbMetas); err != nil {
		slog.Error("failed to encode and write list records response", "error", err)
	}
}
