package model

import (
	"encoding/json"
	"time"
)

type SplitEntity struct {
	ID        string          `gorm:"primaryKey" json:"id"`
	Data      json.RawMessage `gorm:"type:json" json:"data"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}
type Data struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Friends []Friend `json:"friends"`
	Bank    struct {
		Bank            string `json:"bank"`
		BankAccountName string `json:"bankAccountName"`
		BankNumber      string `json:"bankNumber"`
	} `json:"bank"`
	CreatedAt time.Time `json:"createdAt"`
	Bills     []Bill    `json:"bills"`
}

type Friend struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Me        bool         `json:"me"`
	CreatedAt time.Time    `json:"createdAt"`
	Items     []FriendItem `json:"items"`
}

type FriendItem struct {
	BillID    string    `json:"billID"`
	BillName  string    `json:"billName"`
	Price     int       `json:"price"`
	Qty       int       `json:"qty"`
	SubTotal  int       `json:"subTotal"`
	CreatedAt time.Time `json:"createdAt"`
}

type Bill struct {
	ID           string    `json:"id"`
	OwnerID      string    `json:"ownerID"`
	Name         string    `json:"name"`
	Qty          int       `json:"qty"`
	Price        int       `json:"price"`
	CreatedAt    time.Time `json:"createdAt"`
	SplitPayment bool      `json:"splitPayment"`
}
