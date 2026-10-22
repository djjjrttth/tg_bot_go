
package main

import (
	"os"
	"log"
	"github.com/joho/godotenv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


var buttons = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Получить промокод", "get_promo"),
	),

	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Участвовать в розыгрыше", "take_part"),
	),

	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Как участвовать", "get_instruction"),
	),

	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Мои шансы", "get_chances"),
	),

	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Правила акции", "get_rules"),
	),

	
	
)

func handleCallback(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	callback := tgbotapi.NewCallback(query.ID, "")

	if _, err := bot.Request(callback); err != nil {
		log.Panic(err)
	}
	
	chatID := query.Message.Chat.ID

	switch query.Data {
		case "get_promo":
			handleGetPromo(bot, chatID)

		case "take_part":
			handleTakePart(bot, chatID)


	}
}

func handleGetPromo(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Вы нажали на получение промокода")
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func handleTakePart(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Вы нажали на участвовать в розыгрыше")	
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}


func init(){
	if err := godotenv.Load(); err != nil {
		log.Print("File .env not found")
	}
}

func main() {

	api_token, _ := os.LookupEnv("API_TOKEN")

	bot, err := tgbotapi.NewBotAPI(api_token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	
	updateConfig := tgbotapi.NewUpdate(0) // структура UpdateConfig
	updateConfig.Timeout = 30

	// создаем канал с обновлениями
	updates := bot.GetUpdatesChan(updateConfig)


	// пробегаемся по всем обновлениям
	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				switch update.Message.Command(){
				case "start":
					// TODO: добавить приветственный баннер
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать в Бот розыгрышей!")
					msg.ReplyMarkup = buttons
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}

				}
			} 
		} else if update.CallbackQuery != nil {
			handleCallback(bot, update.CallbackQuery)
			continue

			
		} else if update.Message == nil {
			continue
		}

		
	}
}


