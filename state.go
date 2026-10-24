package main

import (
	"time"
)
//var company = "yandex"
//var statusSub bool = false

var prize = Prize {
	name: "айфон 17 про макс",
	instruction: "нажмите на кнопку ниже",
}

var company = Partner {
	name: "Yandex",
	winnersCnt: 3,
	prizeCnt: 3,
	prize: prize,
}

var promo1 = Promo{
	value: "ABCD1234", 
	company: company.name, 
	expiresAt: time.Now().Add(24 * time.Hour),
	issuedAt: time.Now(), 
	usagePlace: "yandex.ru",
}

var promo2 = Promo{
	value: "QWRT45678", 
	company: company.name, 
	expiresAt: time.Now().Add(24 * time.Hour),
	issuedAt: time.Now(), 
	usagePlace: "yandex.ru",
}

var promo3 = Promo{
	value: "ZXCV1234", 
	company: company.name, 
	expiresAt: time.Now().Add(24 * time.Hour),
	issuedAt: time.Now(), 
	usagePlace: "yandex.ru",
}

//var promo2 = Promo{"QWRT45678", company.name, "25.10.2026", "15.10.2026", "yandex.ru"}
//var promo3 = Promo{"ZXCV1234", company.name, "25.10.2026", "15.10.2026", "yandex.ru"}

var part1 = Participant{
	tgID: 10, 
	timestamp : time.Now(),
}

var part2 = Participant{
	tgID: 11, 
	timestamp : time.Now(),
}

var part3 = Participant{
	tgID: 12, 
	timestamp : time.Now(),
}

var part4 = Participant{
	tgID: 13, 
	timestamp : time.Now(),
}

var part5 = Participant{
	tgID: 14, 
	timestamp : time.Now(),
}

var part6 = Participant{
	tgID: 15, 
	timestamp : time.Now(),
}

var part7 = Participant{
	tgID: 16, 
	timestamp : time.Now(),
}

var part8 = Participant{
	tgID: 17, 
	timestamp : time.Now(),
}

var part9 = Participant{
	tgID: 18, 
	timestamp : time.Now(),
}

var part10 = Participant{
	tgID: 19, 
	timestamp : time.Now(),
}



/*var part2 = Participant{11, "11.11.2026", "11:00:00"}
var part3 = Participant{12, "12.12.2026", "12:00:00"}
var part4 = Participant{13, "01.01.2026", "13:00:00"}
var part5 = Participant{14, "02.02.2026", "14:00:00"}
var part6 = Participant{15, "02.02.2026", "14:00:00"}
var part7 = Participant{16, "02.02.2026", "14:00:00"}
var part8 = Participant{17, "02.02.2026", "14:00:00"}
var part9 = Participant{18, "02.02.2026", "14:00:00"}
var part10 = Participant{19, "02.02.2026", "14:00:00"}
*/



var user1 = User{tgID: 10, chances: 1}
var user2 = User{tgID: 11, chances: 10}
var user3 = User{tgID: 12, chances: 7}
var user4 = User{tgID: 13, chances: 2}
var user5 = User{tgID: 14, chances: 5}
var user6 = User{tgID: 15, chances: 4}
var user7 = User{tgID: 16, chances: 3}
var user8 = User{tgID: 17, chances: 7}
var user9 = User{tgID: 18, chances: 10}
var user10 = User{tgID: 19, chances: 1}
var available_promos = []Promo{promo1, promo2, promo3}
var users = map[int64]*User{
	10: &user1, 
	11: &user2, 
	12: &user3, 
	13: &user4, 
	14: &user5, 
	15: &user6, 
	16: &user7, 
	17: &user8, 
	18: &user9, 
	19: &user10, 

}
/*users[10] = &user1
users[11] = &user2
users[12] = &user3
users[13] = &user4
users[14] = &user5*/

var participants = []*Participant{
	&part1, 
	&part2, 
	&part3, 
	&part4, 
	&part5, 
	&part6, 
	&part7, 
	&part8,
	&part9,
	&part10, 

}

var dailyDraws = []Draw{}
var finalDraws = []Draw{}

var pendingDailyDraws = make(map[int64]Draw)
var pendingFinalDraws = make(map[int64]Draw)