package main

import (
	"time"
)

type User struct {

	tgID int64
	username string
	lastname string
	promocode *Promo
	chances uint8
	registered bool
	active bool
	winnerCreds *winnerCreds

}

type Promo struct {
	value string
	company string
	expiresAt time.Time
	issuedAt time.Time
	usagePlace string
}

type Participant struct{
	tgID int64
	timestamp time.Time
}

type Partner struct {
	name string
	winners_cnt uint32
	prizes_cnt uint32
}

type Draw struct {
	timestamp time.Time
	participants []*Participant
	//participants []Participant

	winners []int64
}

type winnerCreds struct {
	firstName string
	lastName  string
	consent   *bool
}