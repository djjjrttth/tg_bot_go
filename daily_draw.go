package main

import (
	"fmt"
	"time"
	"log"
	"math/rand/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
		log.Printf("Недостаточно участников для розыгрыша: %d из %d", participantsCnt, p.WinnersCnt)
		return nil
	}

	winInds := rand.Perm(participantsCnt)
	winnerIDs := make([]int64, 0, p.WinnersCnt)

	for i := uint32(0); i < p.WinnersCnt; i++ {
		winnerIDs = append(winnerIDs, participantTgIDs[winInds[i]])
	}

	return winnerIDs
}

// ... остальные функции (dailyDraw, sendMsgToDailyWinners и т.д.) остаются без изменений ...

func dailyDraw(bot *tgbotapi.BotAPI, p Partner) {
	winnerIDs := getDailyWinnersID(p)
	fmt.Println(winnerIDs)
	
	if winnerIDs != nil {
		newDraw := Draw{
			Timestamp:    time.Now(),      
			Participants: participants,    
			Winners:      winnerIDs,       
			IsFinal:      false,           
		}
		dailyDraws = append(dailyDraws, newDraw)
		
		
		sendMsgToDailyWinners(bot, winnerIDs, p)
		fmt.Println("Ежедневный розыгрыш проведен")
	}
}

func sendMsgToDailyWinners(bot *tgbotapi.BotAPI, dailyWinnerIDs []int64, p Partner) {
	
	text := fmt.Sprintf("Ты в числе хедлайнеров дня 🔥\n" + 
		"Ты выиграл %s от %s.\n" +
		"Чтобы получить приз, %s.",
		p.PrizeName, 
		p.Name, 
		p.PrizeInstruction,
	)

	for _, id := range dailyWinnerIDs {
		if u, ok := users[id]; ok {
			u.PrizeNotifiedAt = time.Now()
			u.IsPrizeClaimed = false
		}
		sendMsgKeyboard(bot, id, text, getPrizeButton)
	}
}

func previewDailyDraw(bot *tgbotapi.BotAPI, adminID int64, p Partner) {
	winnerIDs := getDailyWinnersID(p)
	if winnerIDs == nil {
		sendMsg(bot, adminID, "Недостаточно участников для проведения розыгрыша.")
		return
	}
	
	pendingDailyDraws[adminID] = &PendingDrawData{
		WinnerIDs: winnerIDs,
		Partner:   p,
	}
	
	var winnersText string
	for _, id := range winnerIDs {
		u := users[id]
		name := u.Username 
		if name == "" {
			name = fmt.Sprintf("ID: %d", id)
		}
		winnersText += fmt.Sprintf("• @%s (ID: %d)\n", name, id)
	}
	
	text := fmt.Sprintf("*Предварительные результаты розыгрыша от %s*\n\n"+
		"Победители (%d шт.):\n%s\n\n"+
		"Нажмите кнопку ниже, чтобы подтвердить и отправить уведомления.", 
		p.Name, p.WinnersCnt, winnersText)
	
	sendMsgKeyboard(bot, adminID, text, confirmDailyDrawButton)
}

func confirmAndSendDailyDraw(bot *tgbotapi.BotAPI, adminID int64, p Partner) {
	data, exists := pendingDailyDraws[adminID]
	if !exists {
		sendMsg(bot, adminID, "Ошибка: нет неподтвержденного розыгрыша.")
		return
	}
	
	delete(pendingDailyDraws, adminID)
	
	sendMsgToDailyWinners(bot, data.WinnerIDs, data.Partner)
	
	text := fmt.Sprintf("✅ Розыгрыш от %s успешно подтвержден!\n\n" + 
		"Уведомления отправлены %d победителям.\n" + 
		"Результаты сохранены в историю.", 
		data.Partner.Name, len(data.WinnerIDs))
	
	sendMsg(bot, adminID, text)
}

func startDailyDraw(bot *tgbotapi.BotAPI, adminID int64) {
	go func() {
		for {
			now := time.Now()
			next := now.Add(1 * time.Minute)
			
			// next := time.Date(now.Year(), now.Month(), now.Day()+1, 12, 0, 0, 0, now.Location())
			
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}
			
			pauseDuration := time.Until(next)
			fmt.Printf("Следующий ежедневный розыгрыш через %v\n", pauseDuration)
			time.Sleep(pauseDuration)
			
			previewDailyDraw(bot, adminID, company)
		}
	}()
}