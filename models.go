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

	PrizeNotifiedAt time.Time
	IsPrizeClaimed  bool      

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

type Prize struct {
	name string
	instruction string
}


type Partner struct {
	name string
	winnersCnt uint32
	prizeCnt uint32
	prize Prize
}


type Draw struct {
	timestamp time.Time
	participants []*Participant
	winners []int64        
	reserveWinners []int64 
	isFinal bool           
}

type winnerCreds struct {
	firstName string
	lastName  string
	consent   *bool
}