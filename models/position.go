package models

import "time"

type Position struct {
	ID         uint `gorm:"primaryKey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ClosedAt   *time.Time
	Exchange   string  `json:"exchange"`
	OrderID    string  `gorm:"uniqueIndex;not null"`
	Symbol     string  `gorm:"size:20;not null"`
	Side       string  `gorm:"size:10;not null"`
	Qty        float64 `gorm:"not null"`
	EntryPrice float64 `gorm:"not null"`
	ExitPrice  *float64
	Profit     *float64
	Commission *float64
	Status     string `gorm:"size:10;default:'open'"`
	Reason     string `gorm:"type:text"`
}
