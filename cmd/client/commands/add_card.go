package commands

import (
	"fmt"
	"goph_keeper/internal/shared/models"
)

var (
	cardKey string = name
	cvv     string
	number  string
	expire  string
)

func init() {
	addCmd.Flags().StringVar(&number, "number", "", "set credit card number")
	addCmd.Flags().StringVar(&cvv, "cvv", "", "set credit card cvv")
	addCmd.Flags().StringVar(&expire, "expire", "", "set credit card expired")
}

func handleCard() (models.CardData, error) {
	if cvv == "" || number == "" || expire == "" {
		return models.CardData{}, fmt.Errorf("--cvv,--number, --expire are required for --card")
	}

	cardData := models.CardData{
		CardNumber:     number,
		CVV:            cvv,
		ExpirationDate: expire,
		Name:           name,
	}

	return cardData, nil
}
