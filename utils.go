package main

import (
	"log"
	"strconv"
	"encoding/csv"
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


/*func findParticipantByID(id int64) (int, bool){
	var n int = len(participants)

	for i := range n {
		p := participants[i]
		if p.TgID == id {
			return i, true
		}
	}

	return -1, false
}*/


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

func boolPtr(b bool) *bool {
	return &b
}


func sendMsgMarkdown(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}

func sendMsgKeyboardMarkdown(bot *tgbotapi.BotAPI, chatID int64, text string, buttons ...tgbotapi.InlineKeyboardButton) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttons...),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения с клавиатурой: %v", err)
	}
}

func writeUserRow(writer *csv.Writer, drawID uint, drawType, drawDate, partnerName, prizeName, role string, userID int64) {
	var u User
	status := "Неизвестно"
	username := ""
	firstName := ""
	lastName := ""
	consent := "Нет"

	if err := DB.First(&u, "tg_id = ?", userID).Error; err == nil {
		username = u.Username
		if u.WinnerCreds != nil {
			firstName = u.WinnerCreds.FirstName
			lastName = u.WinnerCreds.LastName
			if u.WinnerCreds.Consent != nil && *u.WinnerCreds.Consent {
				consent = "Да"
			}
		}

		if u.State == "disqualified" {
			status = "Не учавствует"
		} else if u.IsPrizeClaimed {
			status = "Приз получен"
		} else if u.State == "completed" {
			status = "Данные собраны, ожидает передачи"
		} else {
			status = "Ожидает подтверждения"
		}
	} else {
		status = "Пользователь не найден в БД"
	}

	row := []string{
		strconv.FormatUint(uint64(drawID), 10),
		drawType,
		drawDate,
		partnerName,
		prizeName,
		role,
		strconv.FormatInt(userID, 10),
		username,
		firstName,
		lastName,
		consent,
		status,
	}

	if err := writer.Write(row); err != nil {
		log.Printf("Ошибка записи строки CSV для пользователя %d: %v", userID, err)
	}
}