package main

import (
	"log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func sendMsg(bot *tgbotapi.BotAPI, chatID int64, text string){
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}

func sendMsgKeyboard(bot *tgbotapi.BotAPI, chatID int64, text string, button ...tgbotapi.InlineKeyboardButton) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = convertToMarkup(button)

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}


func findParticipantByID(id int64) (int, bool){
	var n int = len(participants)

	for i := range n {
		p := participants[i]
		if p.TgID == id {
			return i, true
		}
	}

	return -1, false
}


func isSubscribed(userID int64, partnerID uint) bool {
	var sub Subscription
	err := DB.Where("tg_id = ? AND partner_id = ?", userID, partnerID).First(&sub).Error
	if err == nil {
		return true
	}
	log.Printf("Подписка не найдена для пользователя %d у партнёра %d", userID, partnerID)
	return false
}


func markSubscribed(userID int64, partnerID uint) error {

	var existing Subscription
	if err := DB.Where("tg_id = ? AND partner_id = ?", userID, partnerID).First(&existing).Error; err == nil {
		return nil
	}

	sub := Subscription{
		TgID:      userID,
		PartnerID: partnerID,
	}
	return DB.Create(&sub).Error
}