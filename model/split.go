package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type SplittedOther struct {
	Name          string  `json:"name"`
	ID            string  `json:"id"`
	Amount        float64 `json:"amount"`
	Price         float64 `json:"price"`
	Type          string  `json:"type"`
	UsePercentage bool    `json:"usePercentage"`
}

func (so SplittedOther) HasFormula() bool {
	return so.Type == "tax"
}

func (so SplittedOther) IsTax() bool {
	return so.Type == "tax"
}

func (so SplittedOther) IsDiscount() bool {
	return so.Type == "deduction" && so.UsePercentage
}

func (so SplittedOther) FormattedPrice() string {
	if so.Type == "deduction" {
		return fmt.Sprintf("-%s", formatCurrency(so.Price))
	}

	return formatCurrency(so.Price)
}

func (so SplittedOther) GetFormula(multiplier float64) string {
	return fmt.Sprintf("%d%% * %s", int64(so.Amount), formatCurrency(multiplier))
}

type SplittedItem struct {
	Name  string  `json:"name"`
	ID    string  `json:"id"`
	Price float64 `json:"price"`
	Equal bool    `json:"equal"`
	Qty   float64 `json:"qty"`
}

func (si SplittedItem) FormattedQty() string {
	return fmt.Sprintf("x%d", int64(si.Qty))
}

func (si SplittedItem) FormattedPrice() string {
	return formatCurrency(si.Price)
}

type SplittedFriend struct {
	Total       float64         `json:"total"`
	Subtotal    float64         `json:"subTotal"`
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
	GrandTotal  float64          `json:"grandTotal"`
	Subtotal    float64          `json:"subTotal"`
}

func (s Splitted) TotalFriends() int {
	return len(s.Friends)
}

func (s Splitted) FormattedCreatedAt() string {
	return s.CreatedAt.Format("02 January 2006")
}

func (s Splitted) EmptyBank() bool {
	return s.BankName == "" && s.BankAccount == "" && s.BankNumber == ""
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

func formatCurrency(amount float64) string {
	return fmt.Sprintf("IDR %s", formatNumber(amount))
}

// Helper function to format number with dot as thousand separator
func formatNumber(n float64) string {
	in := fmt.Sprintf("%d", int64(n))
	out := ""
	for i, c := range in {
		if i > 0 && (len(in)-i)%3 == 0 {
			out += "."
		}
		out += string(c)
	}
	return out
}
