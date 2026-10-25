package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleExport формирует выгрузку со статистикой в формате csv
func handleExport(bot *tgbotapi.BotAPI, chatID int64) {

	var draws []Draw
	if err := DB.Find(&draws).Error; err != nil {
		log.Printf("Ошибка получения розыгрышей для выгрузки: %v", err)
		sendMsg(bot, chatID, "Ошибка при формировании выгрузки.")
		return
	}

	if len(draws) == 0 {
		sendMsg(bot, chatID, "В базе данных пока нет проведенных розыгрышей.")
		return
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)


	headers := []string{
		"ID Розыгрыша", "Тип", "Дата и время", 
		"Партнер", "Название приза", "Роль (Основной/Резерв)", 
		"Telegram ID", "Username", "Имя", "Фамилия", "Согласие на ПДн", "Статус приза",
	}
	if err := writer.Write(headers); err != nil {
		log.Printf("Ошибка записи заголовков CSV: %v", err)
		return
	}

	for _, draw := range draws {
		drawType := "Ежедневный"
		if draw.IsFinal {
			drawType = "Финальный"
		}
		
		partnerName := "Главный приз"
		prizeName := "Билет на Канье Уэста" 
		
		if draw.PartnerID != nil {
			var p Partner
			if err := DB.First(&p, *draw.PartnerID).Error; err == nil {
				partnerName = p.Name
				prizeName = p.PrizeName
			}
		}

		drawDate := draw.Timestamp.Format("2006-01-02 15:04:05")


		for _, winnerID := range draw.Winners {
			writeUserRow(writer, draw.ID, drawType, drawDate, partnerName, prizeName, "Основной", winnerID)
		}

		for _, reserveID := range draw.ReserveWinners {
			writeUserRow(writer, draw.ID, drawType, drawDate, partnerName, prizeName, "Резерв", reserveID)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Printf("Ошибка завершения CSV: %v", err)
		sendMsg(bot, chatID, "Ошибка при формировании файла.")
		return
	}

	fileName := fmt.Sprintf("draws_export_%s.csv", time.Now().Format("20060102_150405"))
	
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{
		Name: fileName,
		Bytes: buf.Bytes(),
	})
	doc.Caption = fmt.Sprintf("Выгрузка результатов розыгрышей.\nВсего розыгрышей: %d\nФормат: CSV", len(draws))
	
	if _, err := bot.Send(doc); err != nil {
		log.Printf("Ошибка отправки файла выгрузки: %v", err)
		sendMsg(bot, chatID, "Не удалось отправить файл. Попробуйте позже.")
	} else {
		log.Printf("Файл выгрузки отправлен админу %d", chatID)
	}
}


