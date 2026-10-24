package main

import (
	"fmt"
	"time"
	"math/rand/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


func getDailyWinnersID(p Partner) []int64 {

	fmt.Println(participants)
	var n uint32 = p.winnersCnt
	var participantsCnt uint32 = uint32(len(participants))
	winnerIDs := make([]int64, 0)
	
	if participantsCnt >= n {
		
		winInds := rand.Perm(int(participantsCnt))

		fmt.Println(winInds)
		fmt.Println()

		for i := range n {
			winnerIDs = append(winnerIDs, participants[winInds[i]].tgID)
		}
		fmt.Println()

		fmt.Println(winnerIDs)
		return winnerIDs

	}

	return nil
}


func dailyDraw(bot *tgbotapi.BotAPI, p Partner){
	winnerIDs := getDailyWinnersID(p)
	fmt.Println(winnerIDs)

	var newDraw = Draw{
		timestamp: time.Now(),
		participants : participants,
		winners : winnerIDs,
	}

	dailyDraws = append(dailyDraws, newDraw)
	sendMsgToDailyWinners(bot, winnerIDs)
	fmt.Println(dailyDraws)
}

func sendMsgToDailyWinners(bot *tgbotapi.BotAPI, dailyWinnerIDs []int64){

	text := fmt.Sprintf("Ты в числе хедлайнеров дня 🔥\n" + 
		"Ты выиграл %s от %s.\n" +
		"Чтобы получить приз, %s.",
		company.prize.name, 
		company.name, 
		company.prize.instruction,
	)

	for _, id := range dailyWinnerIDs {
		sendMsgKeyboard(bot, id, text, getPrizeButton)
	}
}


func previewDailyDraw(bot *tgbotapi.BotAPI, adminID int64, p Partner) {
	winnerIDs := getDailyWinnersID(p)

	if winnerIDs == nil {
		return
	}

	newDraw := Draw{
		timestamp:    time.Now(),
		participants: participants,
		winners:      winnerIDs,
	}
	pendingDailyDraws[adminID] = newDraw

	//TODO: переделать текст
	var winnersText string
	for _, id := range winnerIDs {
		u := users[id]
		name := u.username
		if name == "" {
			name = fmt.Sprintf("ID: %d", id)
		}
		winnersText += fmt.Sprintf("• @%s (ID: %d)\n", name, id)
	}

	text := fmt.Sprintf("*Предварительные результаты розыгрыша от %s*\n\n"+
		"Победители (%d шт.):\n%s\n\n"+
		"Нажмите кнопку ниже, чтобы подтвердить и отправить уведомления.", 
		p.name, p.winnersCnt, winnersText)

	sendMsgKeyboard(bot, adminID, text, confirmDailyDrawButton)
}

func confirmAndSendDailyDraw(bot *tgbotapi.BotAPI, adminID int64, p Partner){
	draw, exists := pendingDailyDraws[adminID]
	if !exists {
		return
	}

	dailyDraws = append(dailyDraws, draw)
	delete(pendingDailyDraws, adminID)
	sendMsgToDailyWinners(bot, draw.winners)

	text := fmt.Sprintf("✅ Розыгрыш от %s успешно подтвержден!\n\n" + 
		"Уведомления отправлены %d победителям.\n" + 
		"Результаты сохранены в историю.", 
		p.name, len(draw.winners),
	)
	
	sendMsg(bot, adminID, text)
}


func startDailyDraw(bot *tgbotapi.BotAPI, adminID int64) {
	go func() {

		for {
			now := time.Now()
			//next := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
			next := now.Add(1 * time.Minute)
			if now.After(next){
				next = next.Add(24 * time.Hour)
			}

			pauseDuration := time.Until(next)
			time.Sleep(pauseDuration)

			previewDailyDraw(bot, adminID, company)
		}
		
	}()
}
