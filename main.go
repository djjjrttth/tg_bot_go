package main

import (
	"os"
	"log"
	"strconv"
	"github.com/joho/godotenv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("File .env not found")
	}
}

func main() {
	api_token, _ := os.LookupEnv("API_TOKEN")
	adminIDstr, _ := os.LookupEnv("ADMIN_ID")
	adminID_, _ := strconv.Atoi(adminIDstr)
	var adminID int64 = int64(adminID_)

	InitDB() 

	bot, err := tgbotapi.NewBotAPI(api_token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false

	startDailyDraw(bot, adminID)
	startPrizeChecker(bot)

	updateConfig := tgbotapi.NewUpdate(0) 
	updateConfig.Timeout = 30
	
	updates := bot.GetUpdatesChan(updateConfig)
	
	for update := range updates {
		if update.Message != nil {
			newUser := update.Message.From
			if update.Message.Text != "" && !update.Message.IsCommand() {
				handleInputCreds(bot, update.Message.Chat.ID, newUser.ID, update.Message.Text)
			} else if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":

					var u User
					DB.FirstOrCreate(&u, User{
						TgID:     newUser.ID,
						Username: newUser.UserName,
						Lastname: newUser.LastName,
					})
					text := "Добро пожаловать в Бот розыгрышей!"
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
					msg.ReplyMarkup = buttons
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}

				case "statistics":
					if newUser.ID == adminID {
        				handleStats(bot, update.Message.Chat.ID)
					} else {
						sendMsg(bot, update.Message.Chat.ID, "Только администраторы могут просматривать статистику")
					}
				
				case "export":
					if newUser.ID == adminID {
        				handleExport(bot, update.Message.Chat.ID)
					} else {
						sendMsg(bot, update.Message.Chat.ID, "Только администраторы могут выгружать статистику")
					}
				case "final":
					if newUser.ID == adminID {
        				startFinalDraw(bot, adminID)
					} else {
						sendMsg(bot, update.Message.Chat.ID, "Только администраторы могут проводить финальные розыгрыши")
					}
				}
			

			} 
		} else if update.CallbackQuery != nil {
			handleCallback(bot, update.CallbackQuery, adminID)
			continue
		} else if update.Message == nil {
			continue
		}
	}
}