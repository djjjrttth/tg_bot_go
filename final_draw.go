package main

import (
	"fmt"
	"log"
	"time"
	"slices"
	"math/rand/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	WinnersCnt = 10
	finalWinnersCnt = 5
)


func chooseWinners(p Partner, uniqueParticipantsCnt int32, participantIDs []int64) []int64 {
	if uniqueParticipantsCnt < WinnersCnt {
		log.Printf("Недостаточно участников для проведения главного розыгрыша")
		return nil
	}

	temp := make([]int64, len(participantIDs))
	copy(temp, participantIDs)

	winners := make([]int64, 0, WinnersCnt)

	
	for len(winners) < WinnersCnt {
		ind := rand.IntN(len(temp))
		winnerID := temp[ind]

		winners = append(winners, winnerID)

		temp = slices.DeleteFunc(temp, func(id int64) bool {
			return id == winnerID
		})

	}

	return winners

}


func getWinnersID(p Partner) ([]int64, []int64) {
	var participantsWithChancesIDs []int64

	
	if len(participants) >= WinnersCnt {
		for _, p := range participants {
			u := users[p.tgID]
			for range u.chances {
				participantsWithChancesIDs = append(participantsWithChancesIDs, p.tgID)
			}

		}

		fmt.Println(participantsWithChancesIDs)

		var finalWinners []int64
		var reserveWinners []int64

		uniqueParticipantsCnt := int32(len(participants))
		winners := chooseWinners(p, uniqueParticipantsCnt, participantsWithChancesIDs)

		finalWinners = winners[:finalWinnersCnt]
		reserveWinners = winners[finalWinnersCnt:]
		

		fmt.Printf("Главные победители: %v\n", finalWinners)
		fmt.Printf("Резервные победители: %v\n\n", reserveWinners)

		return finalWinners, reserveWinners

	}

	log.Printf("Недостаточно участников для проведения финального розыгрыша")
	return nil, nil
}

func runFinalDraw(bot *tgbotapi.BotAPI, p Partner){
	finalWinnerIDs, _ := getWinnersID(p)

	if finalWinnerIDs != nil {
		sendMsgToFinalWinners(bot, finalWinnerIDs)
	}
	

}

func sendMsgToFinalWinners(bot *tgbotapi.BotAPI, finalWinnerIDs []int64){

	text := fmt.Sprintf("Твой сентябрь выходит на главную сцену 🔥\n" + 
		"Ты выиграл билет на Канье Уэста!" + 
		"Чтобы подтвердить получение приза, нажми кнопку ниже и оставь необходимые контактные данные.")

	for _, id := range finalWinnerIDs {
		sendMsgKeyboard(bot, id, text, getPrizeButton)
	}
}


func previewFinalDraw(bot *tgbotapi.BotAPI, adminID int64, p Partner) {

	//fmt.Println("Подтвержаем финальных победителей")
	mainWinners, reserveWinners := getWinnersID(p)

	fmt.Println(mainWinners)
	
	if mainWinners == nil {
		log.Printf("Недостаточно главных участников для проведения розыгрыша")
		return
	}

	if reserveWinners == nil {
		log.Printf("Недостаточно резервных участников для проведения розыгрыша")
		return
	}

	
	pendingFinalDraws[adminID] = Draw{
		timestamp:      time.Now(),
		participants:   participants,
		winners:        mainWinners,
		reserveWinners: reserveWinners,
		isFinal:        true,
	}

	var mainText, reserveText string
	
	for _, id := range mainWinners {
		u := users[id]
		name := u.username
		if name == "" {
			name = fmt.Sprintf("ID: %d", id)
		}
		mainText += fmt.Sprintf("• @%s\n", name)
	}
	
	for _, id := range reserveWinners {
		u := users[id]
		name := u.username
		if name == "" {
			name = fmt.Sprintf("ID: %d", id)
		}
		reserveText += fmt.Sprintf("• @%s\n", name)
	}

	text := fmt.Sprintf("🏆 *ПРЕДВАРИТЕЛЬНЫЕ ИТОГИ ФИНАЛЬНОГО РОЗЫГРЫША*\n\n"+
		"🥇 Главные победители (по 2 билета):\n%s\n"+
		"🥈 Резервные победители:\n%s\n\n"+
		"Нажмите кнопку ниже для подтверждения.", 
		mainText, reserveText)
	
	
	sendMsgKeyboard(bot, adminID, text, confirmFinalDrawButton)
}


func confirmAndSendFinalDraw(bot *tgbotapi.BotAPI, adminID int64) {

	draw, exists := pendingFinalDraws[adminID]
	if !exists {
		sendMsg(bot, adminID, "Ошибка: нет неподтвержденного финального розыгрыша. Возможно, он уже был проведен.")
		return
	}


	finalDraws = append(finalDraws, draw)

	delete(pendingFinalDraws, adminID)

	sendMsgToFinalWinners(bot, draw.winners)

	reserveText := "Ты в резерве финального розыгрыша! 🎟\nЕсли один из главных победителей не заберет свой приз в течение 24 часов, он перейдет к тебе. Ожидай обновления."
	for _, id := range draw.reserveWinners {
		sendMsg(bot, id, reserveText)
	}
	text := fmt.Sprintf("Финальный розыгрыш успешно подтвержден!\n\nУведомления отправлены:\n• %d главным победителям\n• %d резервным победителям", 
		len(draw.winners), len(draw.reserveWinners))
	
	sendMsg(bot, adminID, text)
}


// TODO: реализовать проведение финальных розыгрышей
func startFinalDraw(bot *tgbotapi.BotAPI, adminID int64) {

}