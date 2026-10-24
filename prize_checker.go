package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// --- НАСТРОЙКИ ТАЙМЕРОВ ---
// Для продакшена: 24 * time.Hour
// Для тестов: 1 * time.Minute
const PrizeClaimTimeout = 1 * time.Minute 

// Интервал проверки воркера
// Для продакшена: 1 * time.Hour
// Для тестов: 30 * time.Second
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

	//now := time.Now()

	
	for i := range dailyDraws {
		processDraw(&dailyDraws[i], bot, false)
	}

	for i := range finalDraws {
		processDraw(&finalDraws[i], bot, true)
	}
}

func processDraw(draw *Draw, bot *tgbotapi.BotAPI, isFinal bool) {

	for i, winnerID := range draw.winners {
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

	if len(draw.reserveWinners) == 0 {
		log.Printf("Резервные победители для розыгрыша от %s закончились", draw.timestamp)
		return
	}

	nextReserveID := draw.reserveWinners[0]
	
	draw.reserveWinners = draw.reserveWinners[1:] 

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
			company.prize.name, company.name, company.prize.instruction)
	}

	u.PrizeNotifiedAt = time.Now()
	u.IsPrizeClaimed = false

	draw.winners[failedWinnerIndex] = nextReserveID

	sendMsgKeyboard(bot, nextReserveID, text, getPrizeButton)
	
	log.Printf("Приз передан резервному победителю ID: %d", nextReserveID)
}