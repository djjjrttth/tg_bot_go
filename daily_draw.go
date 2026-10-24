package main

import (
	"fmt"
	"time"
	"math/rand/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func getDailyWinnersID(p Partner) []int64 {
	fmt.Println(participants)
	var n uint32 = p.WinnersCnt
	var participantsCnt uint32 = uint32(len(participants))
	winnerIDs := make([]int64, 0)
	
	if participantsCnt >= n {
		winInds := rand.Perm(int(participantsCnt))
		fmt.Println(winInds)
		fmt.Println()
		for i := range n {
			winnerIDs = append(winnerIDs, participants[winInds[i]].TgID)
		}
		fmt.Println()
		fmt.Println(winnerIDs)
		return winnerIDs
	}
	return nil
}

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