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
	
	for _, p := range participants {
		u := users[p.TgID]
		if u != nil {
			for range u.Chances {
				participantsWithChancesIDs = append(participantsWithChancesIDs, p.TgID)
			}
		}
	}
	
	fmt.Println(participantsWithChancesIDs)
	
	var finalWinners []int64
	var reserveWinners []int64
	uniqueParticipantsCnt := int32(len(participants))
	
	winners := chooseWinners(p, uniqueParticipantsCnt, participantsWithChancesIDs)
	if winners == nil {
		return nil, nil
	}
	
	finalWinners = winners[:finalWinnersCnt]
	reserveWinners = winners[finalWinnersCnt:]
	
	fmt.Printf("Главные победители: %v\n", finalWinners)
	fmt.Printf("Резервные победители: %v\n\n", reserveWinners)
	
	return finalWinners, reserveWinners
}

func runFinalDraw(bot *tgbotapi.BotAPI, p Partner) {
	finalWinnerIDs, _ := getWinnersID(p)
	if finalWinnerIDs != nil {
		sendMsgToFinalWinners(bot, finalWinnerIDs)
	}
}

func sendMsgToFinalWinners(bot *tgbotapi.BotAPI, finalWinnerIDs []int64) {
	text := fmt.Sprintf("Твой сентябрь выходит на главную сцену 🔥\n" +
		"Ты выиграл билет на Канье Уэста!\n" +
		"Чтобы подтвердить получение приза, нажми кнопку ниже и оставь необходимые контактные данные.")
		
	for _, id := range finalWinnerIDs {
		if u, ok := users[id]; ok {
			u.PrizeNotifiedAt = time.Now()
			u.IsPrizeClaimed = false
		}
		sendMsgKeyboard(bot, id, text, getPrizeButton)
	}
}

func previewFinalDraw(bot *tgbotapi.BotAPI, adminID int64, p Partner) {
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
	
	pendingFinalDraws[adminID] = &PendingDrawData{
		WinnerIDs:      mainWinners,
		ReserveWinners: reserveWinners,
		Partner:        p,
	}
	
	var mainText, reserveText string
	for _, id := range mainWinners {
		u := users[id]
		name := u.Username
		if name == "" {
			name = fmt.Sprintf("ID: %d", id)
		}
		mainText += fmt.Sprintf("• @%s\n", name)
	}
	for _, id := range reserveWinners {
		u := users[id]
		name := u.Username
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
	data, exists := pendingFinalDraws[adminID]
	if !exists {
		sendMsg(bot, adminID, "Ошибка: нет неподтвержденного финального розыгрыша.")
		return
	}
	delete(pendingFinalDraws, adminID)
	
	sendMsgToFinalWinners(bot, data.WinnerIDs)
	
	reserveText := "Ты в резерве финального розыгрыша! 🎟\nЕсли один из главных победителей не заберет свой приз в течение 24 часов, он перейдет к тебе. Ожидай обновления."
	for _, id := range data.ReserveWinners {
		sendMsg(bot, id, reserveText)
	}
	
	text := fmt.Sprintf("Финальный розыгрыш успешно подтвержден!\n\nУведомления отправлены:\n• %d главным победителям\n• %d резервным победителям", 
		len(data.WinnerIDs), len(data.ReserveWinners))
	sendMsg(bot, adminID, text)
}

func startFinalDraw(bot *tgbotapi.BotAPI, adminID int64) {
	// TODO: реализовать проведение финальных розыгрышей по расписанию
}