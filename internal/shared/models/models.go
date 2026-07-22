package models

import "context"

type TextData struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

type BinaryData struct {
	Name     string `json:"name"`
	FileName string `json:"file_name"`
	Data     []byte `json:"data"`
}

type CardData struct {
	Name           string `json:"name"`
	CardNumber     string `json:"card_number"`
	ExpirationDate string `json:"expiration_date"`
	CVV            string `json:"cvv"`
	CardholderName string `json:"cardholder_name"`
}

type EncryptedRecord struct {
	Name     string `json:"name"`      // Имя записи (например, "My Visa")
	DataType string `json:"data_type"` // "text", "card" или "file"
	Payload  []byte `json:"payload"`   // ЗАШИФРОВАННЫЕ байты исходной структуры
	Nonce    []byte `json:"nonce"`     // Вектор инициализации для AES-GCM
}

type DecryptedRecord struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
	Data     any    `json:"data"`
}

type RecordMeta struct {
	Name     string `json:"name"`      // Имя секреты (например, "My Visa")
	DataType string `json:"data_type"` // Тип секрета (например, "card")
}

type AuthRequest struct {
	Name         string `json:"name"`
	PasswordHash string `json:"password_hash"`
}

type AuthResponse struct {
	Session string `json:"session"`
}

type ctxAuthKeys string

const (
	tokenContextKey    ctxAuthKeys = "jwt_token"
	userNameContextKey ctxAuthKeys = "user_name"
	cryptedContextKey  ctxAuthKeys = "crypted_key"
)

func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenContextKey, token)
}

func WithUserName(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, userNameContextKey, username)
}

func WithCryptedKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, cryptedContextKey, key)
}

func GetToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenContextKey).(string)
	return token, ok
}

func GetUserName(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(userNameContextKey).(string)
	return name, ok
}

func GetCryptedKey(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(cryptedContextKey).(string)
	return key, ok
}
