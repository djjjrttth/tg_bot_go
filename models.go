package main

import (
	"time"
)


// Структура пользователя
// TgID - id в Телеграмм
// Username и Lastname - имя пользователя в Телеграмм
// Promocode - указатель на полученный промокод
// Chances - количество шансов
// WinnerCreds - имя, фамилия, согласие пользователя, которые он передает при получении приза  
// PrizeNotifiedAt - время уведомления о выигрыше приза
// IsPrizeClaimed - флаг получения приза (получен или нет)
// LastActiveDay - дата последнего активного дня
// State 
//JoinedAt time.Time `gorm:"autoCreateTime"`*/
type User struct {
	TgID int64 `gorm:"primaryKey"`
	Username string
	Lastname string
	Promocode *Promo `gorm:"serializer:json;type:jsonb"` 
	Chances uint8 `gorm:"default:0"`
	WinnerCreds *WinnerCreds `gorm:"serializer:json;type:jsonb"` 
	PrizeNotifiedAt time.Time 
	IsPrizeClaimed bool
	LastActiveDay string `gorm:"type:varchar(10)"` 
	State string `gorm:"type:varchar(32);default:''"`
	JoinedAt time.Time `gorm:"autoCreateTime"`
}

// Структура промокода
// ID - id промокода в базе данных
// Value - значение промокода
// PartnerID - id партнера, который выдал промокод
// ExpiresAt - время, до которого можно использовать промокод
// IssuedAt - время выдачи промокода пользователю
// UsagePlace - место (сайт, приложение и т.д.), где можно использовать промокод
// IssuedTo - кому выдан промокод
// Company - название партнера
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


// Структура участника
// ID - id участника
// TgID - id участника в Телеграмм
// PartnerID - id партнера, в чьем розыгрыше участвует пользователь
// 
type Participant struct {
	ID uint `gorm:"primaryKey"`
	TgID int64 `gorm:"uniqueIndex:uq_daily_participant"`
	PartnerID uint `gorm:"uniqueIndex:uq_daily_participant"`
	Day string `gorm:"type:varchar(10);uniqueIndex:uq_daily_participant"`
	Timestamp time.Time
}

// Структура приза
// Name - название приза
// Instruction - инструкция по получению приза
// Description - описание приза
/*type Prize struct {
	Name string
	Instruction string
	Description string
}*/

// Структура партнера
// ID - id партнера
// Name - название партнера
// WinnersCnt - количество победителей, определяемых партнером
// PrizeCnt - количество призов
// ChannelID - id канала партнера
type Partner struct {
	ID uint `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex"`
	ChannelID int64
	WinnersCnt uint32
	PrizeCnt uint32
	PrizeName  string 
	PrizeInstruction string 
	PromoRequestCount uint32 `gorm:"default:0"`
}


// Структура розыгрыша
// ID - id розыгрыша
// Timestamp - время проведения розыгрыша
// PartnerID - id партнерам, проводящего розыгрыш
// Participants - участники розыгрыша
// Winners - основные победители
// ReserveWinners - резервные победители
// IsFinal - флаг, финальный или ежедневный розыгрыш
type Draw struct {
	ID uint `gorm:"primaryKey"` 
	Timestamp time.Time
	PartnerID *uint
	Participants []*Participant `gorm:"serializer:json;type:jsonb"` 
	Winners []int64 `gorm:"serializer:json;type:jsonb"` 
	ReserveWinners []int64 `gorm:"serializer:json;type:jsonb"` 
	IsFinal bool
}


// Структура контактных данных участника
// FirstName - имя участника
// LastName - фамилия участника
// Consent - флаг, определяет наличие согласия передачи данных
type WinnerCreds struct { 
	FirstName string
	LastName string
	Consent *bool
}

// Структура необходимая для подтверждения розыгрыша администратором
type PendingDrawData struct {
	WinnerIDs []int64
	ReserveWinners []int64
	Partner Partner
}

// Структура подписки
// ID - id подписки
// TgID - id пользователя, который подписался на партнера
// PartnerID - id партнера, на которого подписался пользователь
// CreatedAt - дата подписки
type Subscription struct {
	ID        uint      `gorm:"primaryKey"`
	TgID      int64     `gorm:"uniqueIndex:uq_subscription"`
	PartnerID uint      `gorm:"uniqueIndex:uq_subscription"`
	CreatedAt time.Time
}

// Структура для статистики количества пользователей по дням
type DailyStat struct {
	Day   string
	Count int64
}

// Структура для подсчета остатков промокодов по партнерам
type PromoRemaining struct {
	PartnerID uint
	Count     int64
}


// Структура для подсчета количества участников каждого ежедневного розыгрыша
type PartStat struct {
	Day       string
	PartnerID uint
	Count     int64
}