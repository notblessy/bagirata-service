package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
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
	return fmt.Sprintf("%d%% x %s", int64(so.Amount), formatNumber(multiplier))
}

type SplittedItem struct {
	Name                 string  `json:"name"`
	ID                   string  `json:"id"`
	Price                float64 `json:"price"`
	Equal                bool    `json:"equal"`
	Qty                  float64 `json:"qty"`
	Discount             float64 `json:"discount,omitempty"`
	DiscountIsPercentage bool    `json:"discountIsPercentage,omitempty"`
}

func (si SplittedItem) FormattedQty() string {
	if si.Equal {
		return fmt.Sprintf("%d x %s", int64(si.Price), formatNumber(si.Price/si.Qty))
	}

	if si.Qty > 1 {
		return fmt.Sprintf("%d x %s", int64(si.Qty), formatNumber(si.Price))
	}

	return fmt.Sprintf("%d x", int64(si.Qty))
}

func (si SplittedItem) BaseSubTotal() float64 {
	return si.Price * si.Qty
}

func (si SplittedItem) DiscountAmount() float64 {
	if si.Discount > 0 {
		baseTotal := si.BaseSubTotal()
		if si.DiscountIsPercentage {
			return (si.Discount / 100) * baseTotal
		} else {
			return si.Discount
		}
	}
	return 0
}

func (si SplittedItem) SubTotal() float64 {
	baseTotal := si.BaseSubTotal()
	return baseTotal - si.DiscountAmount()
}

func (si SplittedItem) HasDiscount() bool {
	return si.Discount > 0
}

func (si SplittedItem) FormattedPrice() string {
	if si.HasDiscount() {
		return formatCurrency(si.SubTotal())
	}
	return formatCurrency(si.BaseSubTotal())
}

func (si SplittedItem) FormattedBasePrice() string {
	return formatCurrency(si.BaseSubTotal())
}

func (si SplittedItem) FormattedDiscountAmount() string {
	return formatCurrency(si.DiscountAmount())
}

func (si SplittedItem) FormattedNumber() string {
	return formatNumber(si.Qty)
}

type SplittedFriend struct {
	Total       float64         `json:"total"`
	Subtotal    float64         `json:"subTotal"`
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	FriendID    string          `json:"friendId"`
	Me          bool            `json:"me"`
	CreatedAt   string          `json:"createdAt"`
	AccentColor string          `json:"accentColor"`
	Items       []SplittedItem  `json:"items"`
	Others      []SplittedOther `json:"others"`
}

func (sf SplittedFriend) FormattedTotal() string {
	return formatCurrency(sf.Total)
}

func (sf SplittedFriend) FormattedSubTotal() string {
	return formatCurrency(sf.Subtotal)
}

func (sf SplittedFriend) TotalItemDiscounts() float64 {
	total := 0.0
	for _, item := range sf.Items {
		total += item.DiscountAmount()
	}
	return total
}

func (sf SplittedFriend) FormattedTotalItemDiscounts() string {
	return formatCurrency(sf.TotalItemDiscounts())
}

func (sf SplittedFriend) HasItemDiscounts() bool {
	for _, item := range sf.Items {
		if item.HasDiscount() {
			return true
		}
	}
	return false
}

func (sf SplittedFriend) OriginalItemsSubtotal() float64 {
	total := 0.0
	for _, item := range sf.Items {
		total += item.BaseSubTotal()
	}
	return total
}

func (sf SplittedFriend) FormattedOriginalItemsSubtotal() string {
	return formatCurrency(sf.OriginalItemsSubtotal())
}

func (sf SplittedFriend) DiscountedItemsSubtotal() float64 {
	total := 0.0
	for _, item := range sf.Items {
		total += item.SubTotal()
	}
	return total
}

func (sf SplittedFriend) FormattedDiscountedItemsSubtotal() string {
	return formatCurrency(sf.DiscountedItemsSubtotal())
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
	CreatedAt   string           `json:"createdAt"`
	GrandTotal  float64          `json:"grandTotal"`
	Subtotal    float64          `json:"subTotal"`
}

func (s Splitted) TotalFriends() int {
	return len(s.Friends)
}

func (s Splitted) TotalItemDiscounts() float64 {
	total := 0.0
	for _, friend := range s.Friends {
		total += friend.TotalItemDiscounts()
	}
	return total
}

func (s Splitted) FormattedTotalItemDiscounts() string {
	return formatCurrency(s.TotalItemDiscounts())
}

func (s Splitted) HasItemDiscounts() bool {
	for _, friend := range s.Friends {
		if friend.HasItemDiscounts() {
			return true
		}
	}
	return false
}

func (s Splitted) OriginalSubtotal() float64 {
	total := 0.0
	for _, friend := range s.Friends {
		total += friend.OriginalItemsSubtotal()
	}
	return total
}

func (s Splitted) FormattedOriginalSubtotal() string {
	return formatCurrency(s.OriginalSubtotal())
}

func (s Splitted) DiscountedItemsSubtotal() float64 {
	total := 0.0
	for _, friend := range s.Friends {
		total += friend.DiscountedItemsSubtotal()
	}
	return total
}

func (s Splitted) FormattedDiscountedItemsSubtotal() string {
	return formatCurrency(s.DiscountedItemsSubtotal())
}

func (s Splitted) TotalDiscountedItems() int {
	count := 0
	for _, friend := range s.Friends {
		for _, item := range friend.Items {
			if item.HasDiscount() {
				count++
			}
		}
	}
	return count
}

func (s Splitted) FormattedCreatedAt() string {
	// string to time
	timeDate, err := time.Parse("2006-01-02T15:04:05Z", s.CreatedAt)
	if err != nil {
		logrus.Error(fmt.Errorf("failed to parse time: %w", err))
		return ""
	}

	return timeDate.Format("02 January 2006")
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
