package main

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Интервал, в течение которого можно подтвердить получение приза
const PrizeClaimTimeout = 24 * time.Hour

// Как часто происходит проверка подтверждения получения приза
const CheckerInterval = 1 * time.Hour


// startPrizeChecker запускает проверку
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


// checkExpiredPrizes проверяет подтверждение получения приза в установленный срок
func checkExpiredPrizes(bot *tgbotapi.BotAPI) {
	var draws []Draw
	if err := DB.Find(&draws).Error; err != nil {
		log.Printf("Ошибка получения розыгрышей из БД: %v", err)
		return
	}

	for _, draw := range draws {
		for i, winnerID := range draw.Winners {
			var u User
			if err := DB.First(&u, "tg_id = ?", winnerID).Error; err != nil {
				continue
			}

			if u.IsPrizeClaimed || u.PrizeNotifiedAt.IsZero() || u.State == "disqualified" {
				continue
			}

			if time.Since(u.PrizeNotifiedAt) > PrizeClaimTimeout {
				log.Printf("Победитель %d не забрал приз в срок. Передаем резервному.", winnerID)
				
				DB.Model(&u).Updates(map[string]interface{}{
					"is_prize_claimed": true,
					"state": "disqualified",
				})

				sendMsg(bot, winnerID, "К сожалению, вы не подтвердили получение приза в установленный срок. Приз передан резервному победителю.")
				

				transferToReserve(&draw, i, bot, draw.IsFinal)
			}
		}
	}
}


// transferToReserve передает приз резервному победителю
func transferToReserve(draw *Draw, failedWinnerIndex int, bot *tgbotapi.BotAPI, isFinal bool) {
	if len(draw.ReserveWinners) == 0 {
		log.Printf("Резервные победители для розыгрыша от %s закончились", draw.Timestamp.Format("2006-01-02"))
		return
	}

	nextReserveID := draw.ReserveWinners[0]
	draw.ReserveWinners = draw.ReserveWinners[1:] 
	
	draw.Winners[failedWinnerIndex] = nextReserveID

	if err := DB.Save(draw).Error; err != nil {
		log.Printf("Ошибка обновления розыгрыша в БД при передаче приза: %v", err)
		return
	}

	var u User
	if err := DB.First(&u, "tg_id = ?", nextReserveID).Error; err != nil {
		log.Printf("Ошибка: резервный победитель ID %d не найден в БД.", nextReserveID)
		return 
	}

	now := time.Now()
	DB.Model(&u).Updates(map[string]interface{}{
		"prize_notified_at": now,
		"is_prize_claimed":  false,
	})

	var text string
	if isFinal {
		text = "Поздравляем! 🔥 Вы стали победителем главного розыгрыша, так как предыдущий участник не успел подтвердить приз.\n" +
			"Ты выиграл билет на Канье Уэста!\n" +
			"Чтобы подтвердить получение приза, нажми кнопку ниже."
	} else {
	
		var partnerName, prizeName, instruction string
		if draw.PartnerID != nil {
			var p Partner
			if err := DB.First(&p, "id = ?", *draw.PartnerID).Error; err == nil {
				partnerName = p.Name
				prizeName = p.PrizeName
				instruction = p.PrizeInstruction
			} else {
				partnerName = "Партнера"
				prizeName = "приз"
				instruction = "свяжитесь с поддержкой"
			}
		}
		
		text = fmt.Sprintf("Поздравляем! 🔥 Вы стали победителем дня, так как предыдущий участник не успел подтвердить приз.\n"+
			"Ты выиграл %s от %s.\n"+
			"Чтобы получить приз, %s.",
			prizeName, partnerName, instruction)
	}

	sendMsgKeyboard(bot, nextReserveID, text, getPrizeButton)
	log.Printf("Приз передан резервному победителю ID: %d", nextReserveID)
}