package model

import "time"

type RecognizeRequest struct {
	Model string `json:"model"`
}

type Item struct {
	Name  string `json:"name"`
	Qty   int    `json:"qty"`
	Price int    `json:"price"`
}

type OtherPayment struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Amount        int    `json:"amount"`
	UsePercentage bool   `json:"usePercentage"`
}

type Recognized struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	CreatedAt     time.Time      `json:"createdAt"`
	Items         []Item         `json:"items"`
	OtherPayments []OtherPayment `json:"otherPayments"`
}
