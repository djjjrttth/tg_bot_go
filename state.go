package main

var company = "yandex"
//var statusSub bool = false


var promo1 = Promo{"ABCD1234", "yandex", "25.10.2026", "15.10.2026", "yandex.ru"}
var promo2 = Promo{"QWRT45678", "yandex", "25.10.2026", "15.10.2026", "yandex.ru"}
var promo3 = Promo{"ZXCV1234", "yandex", "25.10.2026", "15.10.2026", "yandex.ru"}

var available_promos = []Promo{promo1, promo2, promo3}
var users = make(map[int64]*User)
var participants = make(map[int64]*Registration)
