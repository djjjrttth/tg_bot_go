package main

import (
	"fmt"
	"log"
	"time"
	"slices"
	"math/rand/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


func chooseMainFinalWinners(p Partner, participantIDs []int64) []int64 {
	var finalWinners []int64
	for len(finalWinners) < 5 {
		ind := rand.IntN(len(participantIDs))
		pID:= participantIDs[ind]
		if !slices.Contains(finalWinners, pID) {
			finalWinners = append(finalWinners, pID)
		}

		fmt.Println(finalWinners)
	}

	return finalWinners
}



func chooseReserveFinalWinners(p Partner, participantIDs, mainFinalWinnerIDs []int64) []int64 {
	var reserveWinners []int64
	
	for len(reserveWinners) < 5 {

		ind := rand.IntN(len(participantIDs))
		pID := participantIDs[ind]


		if !slices.Contains(mainFinalWinnerIDs, pID) && !slices.Contains(reserveWinners, pID) {

			reserveWinners = append(reserveWinners, pID)
			

			participantIDs = slices.DeleteFunc(participantIDs, func(pid int64) bool {
				return pid == pID
			})
		} else {

			participantIDs = slices.DeleteFunc(participantIDs, func(pid int64) bool {
				return pid == pID
			})
		}

		fmt.Println(reserveWinners)
	}

	return reserveWinners
}



func getFinalWinnersID(p Partner) ([]int64, []int64) {
	var participantsWithChancesIDs []int64

	//TODO: сделать проверку что уникальных участников не менее 10
	if len(participants) >= 10 {
		for _, p := range participants {
			u := users[p.tgID]
			for range u.chances {
				participantsWithChancesIDs = append(participantsWithChancesIDs, p.tgID)
			}

		}

		fmt.Println(participantsWithChancesIDs)

		finalWinnersID := chooseMainFinalWinners(p, participantsWithChancesIDs)
		reserveWinnersID := chooseReserveFinalWinners(p, participantsWithChancesIDs, finalWinnersID)

		fmt.Printf("Главные победители: %v\n", finalWinnersID)
		fmt.Printf("Резервные победители: %v\n\n", reserveWinnersID)

		return finalWinnersID, reserveWinnersID

	}

	log.Printf("Недостаточно участников для проведения финального розыгрыша")
	return nil, nil
}

func runFinalDraw(bot *tgbotapi.BotAPI, p Partner){
	finalWinnerIDs, reserveWinners := getFinalWinnersID(p)

	if finalWinnerIDs != nil && reserveWinners != nil {
		sendMsgToFinalWinners(bot, finalWinnerIDs)
	}
	

}

func sendMsgToFinalWinners(bot *tgbotapi.BotAPI, finalWinnerIDs []int64){

	text := fmt.Sprintf("Твой сентябрь выходит на главную сцену 🔥\n" + 
		"Ты выиграл билет на Канье Уэста!" + 
		"Чтобы подтвердить получение приза, нажми кнопку ниже и оставь необходимые контактные данные.")

	for _, id := range finalWinnerIDs {
		sendMsgKeyboard(bot, id, text, getPrizeButton)
		/*msg := tgbotapi.NewMessage(id, text)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Получить приз", "getPrize"),
			),
		)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}*/
	}
}


func previewFinalDraw(bot *tgbotapi.BotAPI, adminID int64, p Partner) {

	//fmt.Println("Подтвержаем финальных победителей")
	mainWinners, reserveWinners := getFinalWinnersID(p)

	fmt.Println(mainWinners)
	
	if mainWinners == nil {
		log.Printf("Недостаточно участников для проведения розыгрыша")
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


// confirmAndSendFinalDraw сохраняет результаты и рассылает уведомления победителям (ТЗ п. 5.2)
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