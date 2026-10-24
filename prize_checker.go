package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


const PrizeClaimTimeout = 1 * time.Minute
const CheckerInterval = 30 * time.Second

var stateMutex sync.Mutex

func startPrizeChecker(bot *tgbotapi.BotAPI) {
	go func() {
		ticker := time.NewTicker(CheckerInterval)
		defer ticker.Stop()

		log.Println("Воркер проверки сроков получения призов запущен.")
		for range ticker.C {
			checkExpiredPrizes(bot)
		}
	}()
}

func checkExpiredPrizes(bot *tgbotapi.BotAPI) {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	for i := range dailyDraws {
		processDraw(&dailyDraws[i], bot, false)
	}
	for i := range finalDraws {
		processDraw(&finalDraws[i], bot, true)
	}
}

func processDraw(draw *Draw, bot *tgbotapi.BotAPI, isFinal bool) {
	for i, winnerID := range draw.Winners {
		u := users[winnerID]
		if u == nil {
			continue
		}

		if u.IsPrizeClaimed || u.PrizeNotifiedAt.IsZero() {
			continue
		}

		if now := time.Now(); now.Sub(u.PrizeNotifiedAt) > PrizeClaimTimeout {
			log.Printf("Победитель %d не забрал приз в срок (%v). Передаем резервному.", winnerID, PrizeClaimTimeout)
			sendMsg(bot, winnerID, "К сожалению, вы не подтвердили получение приза в установленный срок. Приз передан резервному победителю.")
			transferToReserve(draw, i, bot, isFinal)
		}
	}
}

func transferToReserve(draw *Draw, failedWinnerIndex int, bot *tgbotapi.BotAPI, isFinal bool) {
	if len(draw.ReserveWinners) == 0 {
		log.Printf("Резервные победители для розыгрыша от %s закончились", draw.Timestamp.Format("2006-01-02"))
		return
	}

	nextReserveID := draw.ReserveWinners[0]
	draw.ReserveWinners = draw.ReserveWinners[1:] 
	
	u := users[nextReserveID]
	if u == nil {
		log.Printf("Ошибка: резервный победитель ID %d не найден в базе пользователей.", nextReserveID)
		return 
	}

	var text string
	if isFinal {
		text = "Поздравляем! 🔥 Вы стали победителем главного розыгрыша, так как предыдущий участник не успел подтвердить приз.\n" +
			"Ты выиграл билет на Канье Уэста!\n" +
			"Чтобы подтвердить получение приза, нажми кнопку ниже."
	} else {
		text = fmt.Sprintf("Поздравляем! 🔥 Вы стали победителем дня, так как предыдущий участник не успел подтвердить приз.\n" +
			"Ты выиграл %s от %s.\n" +
			"Чтобы получить приз, %s.",
			company.PrizeName, company.Name, company.PrizeInstruction)
	}

	u.PrizeNotifiedAt = time.Now()
	u.IsPrizeClaimed = false
	
	draw.Winners[failedWinnerIndex] = nextReserveID
	
	sendMsgKeyboard(bot, nextReserveID, text, getPrizeButton)
	log.Printf("Приз передан резервному победителю ID: %d", nextReserveID)
}