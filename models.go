package main

type User struct {

	tgID int64
	username string
	lastname string
	promocode *Promo
	chances uint8
	registered bool
	active bool

}


type Promo struct {
	value string
	company string
	expires_at string
	issued_at string
	usage_place string
}

type Registration struct{
	date string
	time string
}