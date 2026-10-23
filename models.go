package main

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
	expires_at string
	issued_at string
	usage_place string
}

type Participant struct{
	tgID int64
	date string
	time string
}

type Partner struct {
	name string
	winners_cnt uint32
	prizes_cnt uint32
}

type Draw struct {
	date string
	time string
	participants []*Participant
	winners []int64
}

type winnerCreds struct {
	firstName string
	lastName  string
	consent   *bool
}