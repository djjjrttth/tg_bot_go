package main

//var company = "yandex"
//var statusSub bool = false

var company = Partner {
	name: "Yandex",
	winners_cnt: 3,
	prizes_cnt: 3,
}

var promo1 = Promo{"ABCD1234", company.name, "25.10.2026", "15.10.2026", "yandex.ru"}
var promo2 = Promo{"QWRT45678", company.name, "25.10.2026", "15.10.2026", "yandex.ru"}
var promo3 = Promo{"ZXCV1234", company.name, "25.10.2026", "15.10.2026", "yandex.ru"}

var part1 = Participant{10, "10.10.2026", "10:00:00"}
var part2 = Participant{11, "11.11.2026", "11:00:00"}
var part3 = Participant{12, "12.12.2026", "12:00:00"}
var part4 = Participant{13, "01.01.2026", "13:00:00"}
var part5 = Participant{14, "02.02.2026", "14:00:00"}


var available_promos = []Promo{promo1, promo2, promo3}
var users = make(map[int64]*User)
var participants = []*Participant{&part1, &part2, &part3, &part4, &part5}

var dailyDraws = []Draw{}