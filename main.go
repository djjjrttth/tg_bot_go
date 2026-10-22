package main

import (
	"os"
	"log"
	"fmt"
	//"time"
	"slices"
	"math/rand/v2"
	"github.com/joho/godotenv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type User struct {

	tgID int64
	username string
	lastname string
	promocode *Promo
	chances uint8
	registered bool
	active bool

}


type Promo struct {
	value string
	company string
	expires_at string
	issued_at string
	usage_place string
}

type Registration struct{
	date string
	time string
}

var company = "yandex"
//var statusSub bool = false


var promo1 = Promo{"ABCD1234", "yandex", "25.10.2026", "15.10.2026", "yandex.ru"}
var promo2 = Promo{"QWRT45678", "yandex", "25.10.2026", "15.10.2026", "yandex.ru"}
var promo3 = Promo{"ZXCV1234", "yandex", "25.10.2026", "15.10.2026", "yandex.ru"}

var available_promos = []Promo{promo1, promo2, promo3}
var users = make(map[int64]*User)
var participants = make(map[int64]*Registration)

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
		tgbotapi.NewInlineKeyboardButtonData("Мои шансы", "getChances"),
	),

	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Правила акции", "get_rules"),
	),

	
	
)

func handleCallback(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, user *tgbotapi.User) {

	//fmt.Println(newUser)
	callback := tgbotapi.NewCallback(query.ID, "")

	if _, err := bot.Request(callback); err != nil {
		log.Panic(err)
	}
	
	chatID := query.Message.Chat.ID

	switch query.Data {
		case "get_promo":
			handleGetPromo(bot, chatID, user)

		case "take_part":
			handleTakePart(bot, chatID)
		case "makeSub":
			handleMakeSub(bot, chatID, user)
		case "checkSub":
			handleCheckSub(bot, chatID, user)
		case "getChances":
			handleGetChances(bot, chatID, user)
			

	}
}

func handleGetPromo(bot *tgbotapi.BotAPI, chatID int64, user *tgbotapi.User) {

	//TODO: добавить ограничение на один промокод по дате
	if len(available_promos) > 0 {

		u := users[user.ID]
		if u.promocode != nil {
			text := "Вы сегодня уже получили промокод"
			msg := tgbotapi.NewMessage(chatID, text)
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}

			return
		}
		
		ind := rand.IntN(len(available_promos))
		promo := available_promos[ind]

		u.promocode = &promo
		available_promos = slices.Delete(available_promos, ind, ind+1)

		if !u.active {
			u.active = true
			if u.chances < 10 {
				u.chances++
			}
		}

		text := fmt.Sprintf("Твой фестивальный бонус от %s готов 🎁\nПромокод: %v\nИспользуй его до %s на %s.\n\nТеперь можешь принять участие в розыгрыше призов дня.", promo.company, promo.value, promo.expires_at, promo.usage_place)
		msg := tgbotapi.NewMessage(chatID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	} else {
		text := fmt.Sprintf("Промокоды от %s уже закончились — их разобрали быстрее всех 🔥\nНо ты все еще можешь подписаться на партнера и принять участие в розыгрыше призов дня.", company)
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Участвовать в розыгрыше", "take_part"),
			),
		)

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}

	}
}


func handleMakeSub(bot *tgbotapi.BotAPI, chatID int64, user *tgbotapi.User){
	//fmt.Println(users[user.ID])
	u := users[user.ID]
	//fmt.Println(u.registered)

	u.registered = true
	//fmt.Printf("Статус подписки: %v\n", statusSub)

	text := fmt.Sprintf("Вы подписались на %s\n", company)
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func handleCheckSub(bot *tgbotapi.BotAPI, chatID int64, user *tgbotapi.User){
	u := users[user.ID]
	if u.registered {
		if _, ok := participants[user.ID]; !ok {
			participants[user.ID] = &Registration{"10.10.2026", "12:00:00"}
			if !u.active {
				u.active = true
				if u.chances < 10{
					u.chances++	
				}
			}
			

			text := fmt.Sprintf("Ты в лайнапе ✅\n\nЗаявка на розыгрыш призов от %s зарегистрирована.\n\nВозвращайся завтра — на сцену выйдет новый партнер.\n\nВсего шансов: %d", company, u.chances)
			msg := tgbotapi.NewMessage(chatID, text)
		
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
				
			}
		} else {
			text := fmt.Sprintf("Ты уже участвуешь в сегодняшнем розыгрыше ✅\n\nРезультаты сообщим после подведения итогов.")
			msg := tgbotapi.NewMessage(chatID, text)
		
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		}
		


	} else {
		text := fmt.Sprintf("Пока не вижу подписку 👀\nПодпишись на аккаунт %s, вернись в бот и попробуй еще раз.", company)
		msg := tgbotapi.NewMessage(chatID, text)

		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Подписаться"), "makeSub"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Проверить еще раз", "checkSub"),
			),
		)


	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
	}
}


func handleTakePart(bot *tgbotapi.BotAPI, chatID int64) {

	text := fmt.Sprintf("Чтобы попасть в розыгрыш призов от %s, подпишись на аккаунт партнера.", company)
	msg := tgbotapi.NewMessage(chatID, text)

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Подписаться на %s", company), "makeSub"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Проверить подписку", "checkSub"),
			),
		)


	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func handleGetChances(bot *tgbotapi.BotAPI, chatID int64, user *tgbotapi.User) {
	u := users[user.ID]

	text := fmt.Sprintf("Твой фестивальный пропуск\nНакоплено шансов: %d из 10\nЧем больше дней розыгрыша ты посетишь, тем выше вероятность выигратьт главный приз.", u.chances)
	msg := tgbotapi.NewMessage(chatID, text)
		
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func createUser(newUser *tgbotapi.User){
	users[newUser.ID] = &User{
		newUser.ID,
		newUser.UserName,
		newUser.LastName,
		nil, 
		0, 
		false,
		false,
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
			if update.Message.IsCommand() {
				switch update.Message.Command(){
				case "start":
					// TODO: добавить приветственный баннер

					
					createUser(newUser)
					//fmt.Println(users[newUser.ID])
					//users[newUser.ID] = &User{newUser.ID, newUser.UserName, newUser.LastName, nil, 0, false}
					
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать в Бот розыгрышей!")
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


