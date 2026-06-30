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
