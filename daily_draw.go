package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func getDailyWinnersID(p Partner) []int64 {
	today := time.Now().Format("2006-01-02")
	
	var participantTgIDs []int64
	err := DB.Model(&Participant{}).
		Where("partner_id = ? AND day = ?", p.ID, today).
		Pluck("tg_id", &participantTgIDs).Error
		
	if err != nil {
		log.Printf("❌ Ошибка получения участников из БД: %v", err)
		return nil
	}

	participantsCnt := len(participantTgIDs)
	if uint32(participantsCnt) < p.WinnersCnt {
		log.Printf("⚠️ Недостаточно участников для розыгрыша: %d из %d", participantsCnt, p.WinnersCnt)
		return nil
	}

	winInds := rand.Perm(participantsCnt)
	winnerIDs := make([]int64, 0, p.WinnersCnt)

	for i := uint32(0); i < p.WinnersCnt; i++ {
		winnerIDs = append(winnerIDs, participantTgIDs[winInds[i]])
	}

	log.Printf("🎲 Выбраны победители: %v", winnerIDs)
	return winnerIDs
}

// dailyDraw вызывается при ручном нажатии "Провести розыгрыш" (минуя подтверждение)
func dailyDraw(bot *tgbotapi.BotAPI, p Partner) {
	log.Printf("⚡ ЗАПУЩЕН МГНОВЕННЫЙ РОЗЫГРЫШ (без подтверждения админом)")
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
		log.Printf("❌ Ошибка сохранения ежедневного розыгрыша в БД: %v", err)
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
	log.Printf("✅ Мгновенный розыгрыш от %s проведен", p.Name)
}

func sendMsgToDailyWinners(bot *tgbotapi.BotAPI, dailyWinnerIDs []int64, p Partner) {
	log.Printf("📤 НАЧАЛО ОТПРАВКИ УВЕДОМЛЕНИЙ ПОБЕДИТЕЛЯМ. Всего: %d чел.", len(dailyWinnerIDs))
	
	text := fmt.Sprintf("Ты в числе хедлайнеров дня 🔥\n"+
		"Ты выиграл %s от %s.\n"+
		"Чтобы получить приз, %s.",
		p.PrizeName, p.Name, p.PrizeInstruction)

	for _, id := range dailyWinnerIDs {
		log.Printf("  -> Отправка сообщения о выигрыше пользователю с ID: %d", id)
		sendMsgKeyboard(bot, id, text, getPrizeButton)
	}
}

// previewDailyDraw вызывается по таймеру или кнопкой "Предварительный просмотр"
func previewDailyDraw(bot *tgbotapi.BotAPI, adminID int64, p Partner) {
	log.Printf("👁 ФОРМИРОВАНИЕ ПРЕВЬЮ РОЗЫГРЫША ДЛЯ АДМИНА (ID: %d)", adminID)
	
	winnerIDs := getDailyWinnersID(p)
	if winnerIDs == nil {
		sendMsg(bot, adminID, "Недостаточно участников для проведения розыгрыша.")
		return
	}
	
	// Сохраняем в память для последующего подтверждения ЭТИМ ЖЕ админом
	pendingDailyDraws[adminID] = &PendingDrawData{
		WinnerIDs: winnerIDs,
		Partner:   p,
	}
	
	var winnersText string
	for _, id := range winnerIDs {
		var u User
		if err := DB.First(&u, "tg_id = ?", id).Error; err == nil && u.Username != "" {
			winnersText += fmt.Sprintf("• @%s (ID: %d)\n", u.Username, id)
		} else {
			winnersText += fmt.Sprintf("• ID: %d\n", id)
		}
	}
	
	text := fmt.Sprintf("*Предварительные результаты розыгрыша от %s*\n\n"+
		"Победители (%d шт.):\n%s\n\n"+
		"Нажмите кнопку ниже, чтобы подтвердить и отправить уведомления.", 
		p.Name, p.WinnersCnt, winnersText)
	
	log.Printf("  -> Отправка ПРЕВЬЮ админу с ID: %d", adminID)
	sendMsgKeyboard(bot, adminID, text, confirmDailyDrawButton)
}

func confirmAndSendDailyDraw(bot *tgbotapi.BotAPI, adminID int64) {
	log.Printf("✅ АДМИН (ID: %d) ПОДТВЕРДИЛ РОЗЫГРЫШ", adminID)
	
	data, exists := pendingDailyDraws[adminID]
	if !exists {
		sendMsg(bot, adminID, "Ошибка: нет неподтвержденного розыгрыша.")
		return
	}
	delete(pendingDailyDraws, adminID)
	
	// Сохраняем подтвержденный розыгрыш в БД
	newDraw := Draw{
		Timestamp: time.Now(),
		Winners:   data.WinnerIDs,
		IsFinal:   false,
	}

	if err := DB.Create(&newDraw).Error; err != nil {
		log.Printf("❌ Ошибка сохранения подтвержденного розыгрыша: %v", err)
		return
	}

	// Обновляем статусы
	now := time.Now()
	for _, id := range data.WinnerIDs {
		DB.Model(&User{}).Where("tg_id = ?", id).Updates(map[string]interface{}{
			"prize_notified_at": now,
			"is_prize_claimed":  false,
		})
	}
	
	// ИСПРАВЛЕНО: fmt.Print -> fmt.Printf
	fmt.Printf("🏆 ID подтвержденных победителей: %v\n", data.WinnerIDs)
	
	// ВОТ ЗДЕСЬ отправляются уведомления победителям, а не админу
	sendMsgToDailyWinners(bot, data.WinnerIDs, data.Partner)
	
	text := fmt.Sprintf("✅ Розыгрыш от %s успешно подтвержден!\n\n"+
		"Уведомления отправлены %d победителям.\n"+
		"Результаты сохранены в историю.", 
		data.Partner.Name, len(data.WinnerIDs))
	
	sendMsg(bot, adminID, text)
}

func startDailyDraw(bot *tgbotapi.BotAPI, adminID int64) {
	go func() {
		for {
			now := time.Now()
			next := now.Add(3 * time.Minute) // Для тестов: 3 минуты
			
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}
			
			pauseDuration := time.Until(next)
			log.Printf("⏳ Следующий автоматический preview розыгрыша через %v", pauseDuration)
			time.Sleep(pauseDuration)
			
			previewDailyDraw(bot, adminID, company)
		}
	}()
}