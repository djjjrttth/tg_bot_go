package main

import (
	"log"
	"strconv"
	"encoding/csv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


// sendMsg отправляет пользователю сообщение  
func sendMsg(bot *tgbotapi.BotAPI, chatID int64, text string){
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}
//sendMsgKeyboard отправляет пользователю сообщение с кнопками
func sendMsgKeyboard(bot *tgbotapi.BotAPI, chatID int64, text string, button ...tgbotapi.InlineKeyboardButton) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = convertToMarkup(button)

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}

// isSubscribed проверяет факт подписки пользователя на партнера
func isSubscribed(userID int64, partnerID uint) bool {
	var sub Subscription
	err := DB.Where("tg_id = ? AND partner_id = ?", userID, partnerID).First(&sub).Error
	if err == nil {
		return true
	}
	log.Printf("Подписка не найдена для пользователя %d у партнёра %d", userID, partnerID)
	return false
}

// markSubscribed создает запись о факте подписки пользователя на партнера
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
// boolPtr возвращает указатель на переменную типа bool
func boolPtr(b bool) *bool {
	return &b
}

// writeUserRow создает строку с данными для выгрузки розыгрыша
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

// convertToRow делает ряд кнопок
func convertToRow(button tgbotapi.InlineKeyboardButton) []tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardRow(button)
}

// convertToMarkup создает меню кнопок
func convertToMarkup(buttons []tgbotapi.InlineKeyboardButton) tgbotapi.InlineKeyboardMarkup {
	k := [][]tgbotapi.InlineKeyboardButton{}

	for _, b := range buttons {
		k = append(k, convertToRow(b))
	}

	return tgbotapi.NewInlineKeyboardMarkup(k...)

}