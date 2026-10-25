package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"slices"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	WinnersCnt      = 10
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
	// 1. Получаем всех уникальных участников из БД
	var uniqueTgIDs []int64
	err := DB.Model(&Participant{}).Distinct("tg_id").Pluck("tg_id", &uniqueTgIDs).Error
	if err != nil {
		log.Printf("Ошибка получения уникальных участников: %v", err)
		return nil, nil
	}

	var participantsWithChancesIDs []int64
	
	// 2. Для каждого уникального участника проверяем его шансы в БД
	for _, tgID := range uniqueTgIDs {
		var u User
		if err := DB.First(&u, "tg_id = ?", tgID).Error; err == nil {
			for i := 0; i < int(u.Chances); i++ {
				participantsWithChancesIDs = append(participantsWithChancesIDs, tgID)
			}
		}
	}
	
	uniqueParticipantsCnt := int32(len(uniqueTgIDs))
	winners := chooseWinners(p, uniqueParticipantsCnt, participantsWithChancesIDs)
	if winners == nil {
		return nil, nil
	}
	
	finalWinners := winners[:finalWinnersCnt]
	reserveWinners := winners[finalWinnersCnt:]
	
	log.Printf("Главные победители: %v", finalWinners)
	log.Printf("Резервные победители: %v", reserveWinners)
	
	return finalWinners, reserveWinners
}

func runFinalDraw(bot *tgbotapi.BotAPI, p Partner) {
	finalWinnerIDs, reserveWinnerIDs := getWinnersID(p)
	if finalWinnerIDs == nil {
		return
	}

	// Сохраняем в БД
	newDraw := Draw{
		Timestamp:      time.Now(),
		Winners:        finalWinnerIDs,
		ReserveWinners: reserveWinnerIDs,
		IsFinal:        true,
	}

	if err := DB.Create(&newDraw).Error; err != nil {
		log.Printf("Ошибка сохранения финального розыгрыша: %v", err)
		return
	}

	// Обновляем статусы главных победителей
	now := time.Now()
	for _, id := range finalWinnerIDs {
		DB.Model(&User{}).Where("tg_id = ?", id).Updates(map[string]interface{}{
			"prize_notified_at": now,
			"is_prize_claimed":  false,
		})
	}

	sendMsgToFinalWinners(bot, finalWinnerIDs)
}

func sendMsgToFinalWinners(bot *tgbotapi.BotAPI, finalWinnerIDs []int64) {
	text := fmt.Sprintf("Твой сентябрь выходит на главную сцену 🔥\n" +
		"Ты выиграл билет на Канье Уэста!\n" +
		"Чтобы подтвердить получение приза, нажми кнопку ниже и оставь необходимые контактные данные.")
		
	for _, id := range finalWinnerIDs {
		sendMsgKeyboard(bot, id, text, getPrizeButton)
	}
}

func previewFinalDraw(bot *tgbotapi.BotAPI, adminID int64, p Partner) {
	mainWinners, reserveWinners := getWinnersID(p)
	if mainWinners == nil || reserveWinners == nil {
		sendMsg(bot, adminID, "Недостаточно участников для проведения финального розыгрыша.")
		return
	}
	
	pendingFinalDraws[adminID] = &PendingDrawData{
		WinnerIDs:      mainWinners,
		ReserveWinners: reserveWinners,
		Partner:        p,
	}
	
	var mainText, reserveText string
	
	// Получаем имена из БД
	for _, id := range mainWinners {
		var u User
		if err := DB.First(&u, "tg_id = ?", id).Error; err == nil && u.Username != "" {
			mainText += fmt.Sprintf("• @%s\n", u.Username)
		} else {
			mainText += fmt.Sprintf("• ID: %d\n", id)
		}
	}
	for _, id := range reserveWinners {
		var u User
		if err := DB.First(&u, "tg_id = ?", id).Error; err == nil && u.Username != "" {
			reserveText += fmt.Sprintf("• @%s\n", u.Username)
		} else {
			reserveText += fmt.Sprintf("• ID: %d\n", id)
		}
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
	
	// Сохраняем в БД
	newDraw := Draw{
		Timestamp:      time.Now(),
		Winners:        data.WinnerIDs,
		ReserveWinners: data.ReserveWinners,
		IsFinal:        true,
	}

	if err := DB.Create(&newDraw).Error; err != nil {
		log.Printf("Ошибка сохранения подтвержденного финального розыгрыша: %v", err)
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

	fmt.Println("Начинаем финальный розыгрыш")
}