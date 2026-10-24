package main

import (
	"time"
)

var company = Partner{
	ID:               1,
	Name:             "Yandex",
	ChannelID:  -1001009185002,
	WinnersCnt:       3,
	PrizeCnt:         3,
	PrizeName:        "Айфон 17 Про Макс",
	PrizeInstruction: "перейдите в раздел 'Бонусы' в личном кабинете",
}


var promo1 = Promo{
	ID:         1,
	Value:      "ABCD1234",
	PartnerID:  1,
	ExpiresAt:  time.Now().Add(24 * time.Hour),
	UsagePlace: "yandex.ru",
	IssuedAt:   nil, 
}

var promo2 = Promo{
	ID:         2,
	Value:      "QWRT4567",
	PartnerID:  1,
	ExpiresAt:  time.Now().Add(24 * time.Hour),
	UsagePlace: "yandex.ru",
	IssuedAt:   nil, 
}

var promo3 = Promo{ 
	ID:         3,
	Value:      "ZXCV1234",
	PartnerID:  1,
	ExpiresAt:  time.Now().Add(24 * time.Hour),
	UsagePlace: "yandex.ru",
	IssuedAt:   nil, 
}

var available_promos = []Promo{promo1, promo2, promo3}


var user1 = User{TgID: 10, Chances: 1}
var user2 = User{TgID: 11, Chances: 10}
var user3 = User{TgID: 12, Chances: 7}
var user4 = User{TgID: 13, Chances: 2}
var user5 = User{TgID: 14, Chances: 5}
var user6 = User{TgID: 15, Chances: 4}
var user7 = User{TgID: 16, Chances: 3}
var user8 = User{TgID: 17, Chances: 7}
var user9 = User{TgID: 18, Chances: 10}
var user10 = User{TgID: 19, Chances: 1}

var users = map[int64]*User{
	10: &user1, 11: &user2, 12: &user3, 13: &user4, 14: &user5,
	15: &user6, 16: &user7, 17: &user8, 18: &user9, 19: &user10,
}

var participants = []*Participant{}

var pendingDailyDraws = make(map[int64]*PendingDrawData)
var pendingFinalDraws = make(map[int64]*PendingDrawData)

var dailyDraws = []Draw{}
var finalDraws = []Draw{}