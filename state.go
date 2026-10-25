package main


var company = Partner{
	ID:               1,
	Name:             "Yandex",
	WinnersCnt:       3,
	PrizeCnt:         3,
	PrizeName:        "Айфон 17 Про Макс",
	PrizeInstruction: "перейдите в раздел 'Бонусы' в личном кабинете",
}


var pendingDailyDraws = make(map[int64]*PendingDrawData)
var pendingFinalDraws = make(map[int64]*PendingDrawData)

