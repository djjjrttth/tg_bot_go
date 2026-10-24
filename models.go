package main

import (
	"time"
)

type User struct {
	TgID int64 `gorm:"primaryKey"`
	Username string
	Lastname string
	Promocode *Promo `gorm:"serializer:json;type:jsonb"` 
	Chances uint8 `gorm:"default:0"`
	Registered bool `gorm:"default:false"`
	Active bool `gorm:"default:false"`
	WinnerCreds *WinnerCreds `gorm:"serializer:json;type:jsonb"` 
	PrizeNotifiedAt time.Time 
	IsPrizeClaimed  bool      
}

type Promo struct {
	ID uint `gorm:"primaryKey"`
	Value string `gorm:"uniqueIndex"`
	PartnerID uint `gorm:"index"`
	Company string
	ExpiresAt time.Time
	IssuedAt *time.Time `gorm:"index"` 
	UsagePlace string
}

type Participant struct {
	TgID int64`gorm:"uniqueIndex:uq_daily_participant"`
	Timestamp time.Time
}

type Prize struct {
	Name string
	Instruction string
}

type Partner struct {
	ID uint `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex"`
	ChannelUsername string
	WinnersCnt uint32
	PrizeCnt uint32
	PrizeName  string 
	PrizeInstruction string 
}

type Draw struct {
	ID uint `gorm:"primaryKey"` 
	Timestamp time.Time

	Participants []*Participant `gorm:"serializer:json;type:jsonb"` 
	Winners []int64 `gorm:"serializer:json;type:jsonb"` 
	ReserveWinners []int64 `gorm:"serializer:json;type:jsonb"` 
	IsFinal bool
}

type WinnerCreds struct { 
	FirstName string
	LastName string
	Consent *bool
}

type PendingDrawData struct {
	WinnerIDs []int64
	ReserveWinners []int64
	Partner Partner
}