package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


// getDailyWinnersID случайно выбирает необходимое число победителей
func getDailyWinnersID(p Partner) []int64 {
	today := time.Now().Format("2006-01-02")
	
	var participantTgIDs []int64
	err := DB.Model(&Participant{}).
		Where("partner_id = ? AND day = ?", p.ID, today).
		Pluck("tg_id", &participantTgIDs).Error
		
	if err != nil {
		log.Printf("Ошибка получения участников из БД: %v", err)
		return nil
	}

	participantsCnt := len(participantTgIDs)
	if uint32(participantsCnt) < p.WinnersCnt {
		log.Printf("Недостаточно участников для ежедневного розыгрыша: %d из %d", participantsCnt, p.WinnersCnt)
		return nil
	}

	winInds := rand.Perm(participantsCnt)
	winnerIDs := make([]int64, 0, p.WinnersCnt)

	for i := uint32(0); i < p.WinnersCnt; i++ {
		winnerIDs = append(winnerIDs, participantTgIDs[winInds[i]])
	}

	log.Printf("Выбраны победители: %v", winnerIDs)
	return winnerIDs
}

// dailyDraw вызывается при ручном нажатии "Провести розыгрыш"
func dailyDraw(bot *tgbotapi.BotAPI, p Partner) {
	log.Printf("Запущен ежедневный розыгрыш.")
	winnerIDs := getDailyWinnersID(p)
	if winnerIDs == nil {
		return
	}
	
	newDraw := Draw{
		Timestamp: time.Now(),
		Winners:   winnerIDs,
		IsFinal:   false,
	}

	if err := DB.Create(&newDraw).Error; err != nil {
		log.Printf("Ошибка сохранения ежедневного розыгрыша в БД: %v", err)
		return
	}

	now := time.Now()
	for _, id := range winnerIDs {
		DB.Model(&User{}).Where("tg_id = ?", id).Updates(map[string]interface{}{
			"prize_notified_at": now,
			"is_prize_claimed":  false,
		})
	}

	sendMsgToDailyWinners(bot, winnerIDs, p)
	log.Printf("Ежедневный розыгрыш от %s проведен", p.Name)
}


// sendMsgToDailyWinners отправляет уведомления победителям ежедневного розыгрыша
func sendMsgToDailyWinners(bot *tgbotapi.BotAPI, dailyWinnerIDs []int64, p Partner) {
	log.Printf("Отправление уведомлений победителям", len(dailyWinnerIDs))
	
	text := fmt.Sprintf("Ты в числе хедлайнеров дня 🔥\n"+
		"Ты выиграл %s от %s.\n"+
		"Чтобы получить приз, %s.",
		p.PrizeName, p.Name, p.PrizeInstruction)

	for _, id := range dailyWinnerIDs {
		log.Printf("Отправка сообщения о выигрыше пользователю с ID: %d", id)
		sendMsgKeyboard(bot, id, text, getPrizeButton)
	}
}

// previewDailyDraw вызывается по таймеру
func previewDailyDraw(bot *tgbotapi.BotAPI, adminID int64, p Partner) {
	
	winnerIDs := getDailyWinnersID(p)
	if winnerIDs == nil {
		sendMsg(bot, adminID, "Недостаточно участников для проведения ежедневного розыгрыша.")
		return
	}
	
	// Сохранение в память для последующего подтверждения тем же самым админом
	pendingDailyDraws[adminID] = &PendingDrawData{
		WinnerIDs: winnerIDs,
		Partner:   p,
	}
	
	var winnersText string
	for _, id := range winnerIDs {
		var u User
		if err := DB.First(&u, "tg_id = ?", id).Error; err == nil && u.Username != "" {
			winnersText += fmt.Sprintf("- @%s (ID: %d)\n", u.Username, id)
		} else {
			winnersText += fmt.Sprintf("- ID: %d\n", id)
		}
	}
	
	text := fmt.Sprintf("*Предварительные результаты розыгрыша от %s*\n\n"+
		"Победители (%d шт.):\n%s\n\n"+
		"Нажмите кнопку ниже, чтобы подтвердить и отправить уведомления.", 
		p.Name, p.WinnersCnt, winnersText)
	
	sendMsgKeyboard(bot, adminID, text, confirmDailyDrawButton)
}

// confirmAndSendDailyDraw отправляет подтверждение администратору о победителях и рассылает уведомления победителям
func confirmAndSendDailyDraw(bot *tgbotapi.BotAPI, adminID int64) {
	
	data, exists := pendingDailyDraws[adminID]
	if !exists {
		sendMsg(bot, adminID, "Ошибка: нет неподтвержденного розыгрыша.")
		return
	}
	delete(pendingDailyDraws, adminID)
	

	newDraw := Draw{
		Timestamp: time.Now(),
		Winners:   data.WinnerIDs,
		IsFinal:   false,
	}

	if err := DB.Create(&newDraw).Error; err != nil {
		log.Printf("Ошибка сохранения подтвержденного розыгрыша: %v", err)
		return
	}

	now := time.Now()
	for _, id := range data.WinnerIDs {
		DB.Model(&User{}).Where("tg_id = ?", id).Updates(map[string]interface{}{
			"prize_notified_at": now,
			"is_prize_claimed":  false,
		})
	}
	
	sendMsgToDailyWinners(bot, data.WinnerIDs, data.Partner)
	
	text := fmt.Sprintf("Розыгрыш от %s успешно подтвержден!\n\n"+
		"Уведомления отправлены %d победителям.\n"+
		"Результаты сохранены в историю.", 
		data.Partner.Name, len(data.WinnerIDs))
	
	sendMsg(bot, adminID, text)
}


// Запускает ежедневный розыгрыш
// ежедневный розыгрыш проводится в 12:00
func startDailyDraw(bot *tgbotapi.BotAPI, adminID int64) {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 12, 0, 0, 0, now.Location())
			
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}
			
			pauseDuration := time.Until(next)
			log.Printf("Следующий ежедневный розыгрыш через %v", pauseDuration)
			time.Sleep(pauseDuration)
			
			previewDailyDraw(bot, adminID, company)
		}
	}()
}