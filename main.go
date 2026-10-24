package main

import (
	"os"
	"log"
	//"time"
	"github.com/joho/godotenv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


func createUser(newUser *tgbotapi.User){
	
	users[newUser.ID] = &User{
		tgID: newUser.ID,
		username: newUser.UserName,
		lastname: newUser.LastName,
		promocode: nil, 
		chances: 0, 
		registered: false,
		active: false,
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

	bot.Debug = false
	
	updateConfig := tgbotapi.NewUpdate(0) // структура UpdateConfig
	updateConfig.Timeout = 30

	// создаем канал с обновлениями
	updates := bot.GetUpdatesChan(updateConfig)


	// пробегаемся по всем обновлениям
	for update := range updates {
		
		
		if update.Message != nil {
			
			newUser := update.Message.From
			if update.Message.Text != "" && !update.Message.IsCommand() {
				handleInputCreds(bot, update.Message.Chat.ID, newUser, update.Message.Text)
			} else if update.Message.IsCommand() {
				switch update.Message.Command(){
				case "start":
					// TODO: добавить приветственный баннер

					
					createUser(newUser)
					
					// TODO: вызывать функцию
					text := "Добро пожаловать в Бот розыгрышей!"
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
					msg.ReplyMarkup = buttons
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}

				}
			} 
		} else if update.CallbackQuery != nil {
			user := update.CallbackQuery.From
			handleCallback(bot, update.CallbackQuery, user)
			continue

			
		} else if update.Message == nil {
			continue
		}

		
	}
}


