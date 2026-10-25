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
	IsPrizeClaimed bool
	LastActiveDay string `gorm:"type:varchar(10)"` 
	State string `gorm:"type:varchar(32);default:''"`
	JoinedAt time.Time `gorm:"autoCreateTime"`
}

type Promo struct {
	ID uint `gorm:"primaryKey"`
	Value string `gorm:"uniqueIndex"`
	PartnerID uint `gorm:"index"`
	Company string
	ExpiresAt time.Time
	IssuedAt *time.Time `gorm:"index"` 
	UsagePlace string
	IssuedTo *int64 `gorm:"index"`
	IssuedDay string `gorm:"type:varchar(10);index"`
}

type Participant struct {
	ID uint `gorm:"primaryKey"`
	TgID int64 `gorm:"uniqueIndex:uq_daily_participant"`
	PartnerID uint `gorm:"uniqueIndex:uq_daily_participant"`
	Day string `gorm:"type:varchar(10);uniqueIndex:uq_daily_participant"`
	Timestamp time.Time
}

type Prize struct {
	Name string
	Instruction string
}

type Partner struct {
	ID uint `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex"`
	ChannelID int64
	WinnersCnt uint32
	PrizeCnt uint32
	PrizeName  string 
	PrizeInstruction string 
}

type Draw struct {
	ID uint `gorm:"primaryKey"` 
	Timestamp time.Time
	PartnerID *uint
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

type Subscription struct {
	ID        uint      `gorm:"primaryKey"`
	TgID      int64     `gorm:"uniqueIndex:uq_subscription"`
	PartnerID uint      `gorm:"uniqueIndex:uq_subscription"`
	CreatedAt time.Time
}


type DailyStat struct {
	Day   string
	Count int64
}

type PromoRemaining struct {
	PartnerID uint
	Count     int64
}

type PartStat struct {
	Day       string
	PartnerID uint
	Count     int64
}