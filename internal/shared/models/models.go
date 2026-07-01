package models

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
