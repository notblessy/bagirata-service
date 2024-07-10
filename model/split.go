package model

import (
	"encoding/json"
	"time"
)

type SplittedOther struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Amount int    `json:"amount"`
	Type   string `json:"type"`
}

type SplittedItem struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Price int    `json:"price"`
	Qty   int    `json:"qty"`
}

type SplittedFriend struct {
	Total       int             `json:"total"`
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	FriendID    string          `json:"friendId"`
	Me          bool            `json:"me"`
	CreatedAt   time.Time       `json:"createdAt"`
	AccentColor string          `json:"accentColor"`
	Items       []SplittedItem  `json:"items"`
	Others      []SplittedOther `json:"others"`
}

type Splitted struct {
	Friends    []SplittedFriend `json:"friends"`
	Slug       string           `json:"slug"`
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	CreatedAt  time.Time        `json:"createdAt"`
	GrandTotal int              `json:"grandTotal"`
}

func (s Splitted) ToData() SplitEntity {
	data, _ := json.Marshal(s)

	return SplitEntity{
		ID:   s.ID,
		Slug: s.Slug,
		Data: data,
	}
}

type SplitEntity struct {
	ID   string          `json:"id" gorm:"primaryKey"`
	Slug string          `json:"slug" gorm:"unique"`
	Data json.RawMessage `json:"data" gorm:"type:jsonb"`
}

func (s SplitEntity) TableName() string {
	return "splits"
}
