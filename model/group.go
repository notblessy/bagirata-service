package model

import (
	"time"

	"gorm.io/gorm"
)

// Group represents a named collection of splits (e.g. "Trip to Japan").
type Group struct {
	ID           string         `json:"id" gorm:"primaryKey;type:uuid"`
	UserID       string         `json:"userId" gorm:"type:uuid;index;not null"`
	Name         string         `json:"name" gorm:"not null"`
	BankName     string         `json:"bankName" gorm:"column:bank_name"`
	BankAccount  string         `json:"bankAccount" gorm:"column:bank_account"`
	BankNumber   string         `json:"bankNumber" gorm:"column:bank_number"`
	ShareSlug    *string        `json:"shareSlug,omitempty" gorm:"column:share_slug;uniqueIndex"`
	CreatedAt    time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Group) TableName() string {
	return "groups"
}

// CreateGroupRequest is the body for POST /v1/groups
type CreateGroupRequest struct {
	Name        string `json:"name"`
	BankName    string `json:"bankName"`
	BankAccount string `json:"bankAccount"`
	BankNumber  string `json:"bankNumber"`
}

// UpdateGroupRequest is the body for PATCH /v1/groups/:id
type UpdateGroupRequest struct {
	Name             *string `json:"name"`
	BankName         *string `json:"bankName"`
	BankAccount      *string `json:"bankAccount"`
	BankNumber       *string `json:"bankNumber"`
	GenerateShareSlug *bool   `json:"generateShareSlug"`
}

// GroupSummaryParticipant is one row in GET /v1/groups/:id/summary (aggregated per participant)
type GroupSummaryParticipant struct {
	FriendID    string  `json:"friendId"`
	Name        string  `json:"name"`
	AccentColor string  `json:"accentColor"`
	Me          bool    `json:"me"`
	Total       float64 `json:"total"`
}

// GroupSummarySplit is a split that belongs to the group (for context in summary)
type GroupSummarySplit struct {
	ID           string  `json:"id"`
	Slug         string  `json:"slug"`
	Name         string  `json:"name"`
	GrandTotal   float64 `json:"grandTotal"`
	CreatedAt    string  `json:"createdAt"`
	CurrencyCode string  `json:"currencyCode,omitempty"`
}

// GroupSummaryResponse is the response for GET /v1/groups/:id/summary
type GroupSummaryResponse struct {
	ID           string                    `json:"id"`
	Name         string                    `json:"name"`
	BankName     string                    `json:"bankName"`
	BankAccount  string                    `json:"bankAccount"`
	BankNumber   string                    `json:"bankNumber"`
	ShareSlug    string                    `json:"shareSlug"`
	Participants []GroupSummaryParticipant `json:"participants"`
	Splits       []GroupSummarySplit       `json:"splits"`
}
