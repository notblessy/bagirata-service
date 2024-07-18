package model

import (
	"encoding/json"
	"fmt"
	"strings"
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

func (sf SplittedFriend) FormattedTotal() string {
	return formatCurrency(sf.Total)
}

func (sf SplittedFriend) InitialName() string {
	words := strings.Fields(sf.Name)
	if len(words) >= 2 {
		return strings.ToUpper(string(words[0][0])) + strings.ToUpper(string(words[1][0]))
	} else if len(words) == 1 {
		if len(words[0]) >= 2 {
			return strings.ToUpper(string(words[0][0])) + strings.ToUpper(string(words[0][1]))
		}
		return strings.ToUpper(string(words[0][0]))
	}
	return ""
}

type Splitted struct {
	Friends     []SplittedFriend `json:"friends"`
	Slug        string           `json:"slug"`
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	BankName    string           `json:"bankName"`
	BankAccount string           `json:"bankAccount"`
	BankNumber  string           `json:"bankNumber"`
	CreatedAt   time.Time        `json:"createdAt"`
	GrandTotal  int              `json:"grandTotal"`
}

func (s Splitted) FormattedCreatedAt() string {
	return s.CreatedAt.Format("02 January 2006")
}

// Should have method that returns grand total to "IDR 100.000"
// format also int 100000 to "IDR 100.000"
func (s Splitted) FormattedGrandTotal() string {
	return formatCurrency(s.GrandTotal)
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

func formatCurrency(amount int) string {
	return fmt.Sprintf("IDR %s", formatNumber(amount))
}

// Helper function to format number with dot as thousand separator
func formatNumber(n int) string {
	in := fmt.Sprintf("%d", n)
	out := ""
	for i, c := range in {
		if i > 0 && (len(in)-i)%3 == 0 {
			out += "."
		}
		out += string(c)
	}
	return out
}
