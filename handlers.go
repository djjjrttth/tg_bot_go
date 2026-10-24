package main

import (
	"fmt"
	"log"
	"time"
	"slices"
	"math/rand/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


func handleCallback(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {

	callback := tgbotapi.NewCallback(query.ID, "")

	if _, err := bot.Request(callback); err != nil {
		log.Panic(err)
	}
	
	chatID := query.Message.Chat.ID
	userID := query.From.ID

	switch query.Data {
		case "getPromo":
			handleGetPromo(bot, chatID, userID)

		case "takePart":
			handleTakePart(bot, chatID)
		case "makeSub":
			handleMakeSub(bot, chatID, userID)
		case "checkSub":
			handleCheckSub(bot, chatID, userID)
		case "getChances":
			handleGetChances(bot, chatID, userID)
		case "getInstruction":
			handleGetInstruction(bot, chatID)
		case "getRules":
			handleGetRules(bot, chatID)
		case "getDailyWinners":
			getDailyWinnersID(company)
		case "runDailyDraw":
			dailyDraw(bot, company)
		case "getPrize":
			handleGetPrize(bot, chatID, userID)

		case "runFinalDraw":
			runFinalDraw(bot, company)

		case "previewDailyDraw":
			previewDailyDraw(bot, userID, company)
		case "confirmDailyDraw":
			confirmAndSendDailyDraw(bot, userID, company)
		case "previewFinalDraw":
			previewFinalDraw(bot, userID, company)
			// TODO: Добавить проверку user.ID == AdminID
			
		case "confirmFinalDraw":
			confirmAndSendFinalDraw(bot, userID)
	
	}
}


func handleGetPromo(bot *tgbotapi.BotAPI, chatID int64, userID int64) {

	stateMutex.Lock()
	defer stateMutex.Unlock()

	u := users[userID]
	
	if u.promocode != nil {
		todayDate := time.Now().Format("2006-01-02")
		issuedAtDate := u.promocode.issuedAt.Format("2006-01-02")
		if todayDate == issuedAtDate{
			text := fmt.Sprintf("Вы сегодня уже получили промокод\n\nВаш промокод от %s: %v", u.promocode.company, u.promocode.value)
			sendMsg(bot, chatID, text)
			return
		}
		
		u.promocode = nil

	} 
	if len(available_promos) > 0 {
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

		text := fmt.Sprintf("Твой фестивальный бонус от %s готов 🎁\n" + 
			"Промокод: %v\nИспользуй его до %s на %s.\n\n" +
			"Теперь можешь принять участие в розыгрыше призов дня.", promo.company, promo.value, promo.expiresAt, promo.usagePlace)
		
		sendMsgKeyboard(bot, chatID, text, takePartButton)
		
		
		} else {
			text := fmt.Sprintf("Промокоды от %s уже закончились — их разобрали быстрее всех 🔥\nНо ты все еще можешь подписаться на партнера и принять участие в розыгрыше призов дня.", company.name)
			sendMsgKeyboard(bot, chatID, text, takePartButton)

		}
}

func handleMakeSub(bot *tgbotapi.BotAPI, chatID int64, userID int64){

	stateMutex.Lock()
	defer stateMutex.Unlock()

	u := users[userID]

	u.registered = true

	text := fmt.Sprintf("Вы подписались на %s\n", company.name)
	sendMsg(bot, chatID, text)
}

func handleCheckSub(bot *tgbotapi.BotAPI, chatID int64, userID int64){

	stateMutex.Lock()
	defer stateMutex.Unlock()

	u := users[userID]
	if u == nil {
		log.Printf("Ошибка: пользователя не существует")
		return
	}

	if u.registered {
		if _, ok := findParticipantByID(userID); !ok {
			participants = append(participants, &Participant{
				tgID: userID, 
				timestamp : time.Now(),
			})
			
			if !u.active {
				u.active = true
				if u.chances < 10{
					u.chances++	
				}
			}
			

			text := fmt.Sprintf("Ты в лайнапе ✅\n\nЗаявка на розыгрыш призов от %s зарегистрирована.\n\nВозвращайся завтра — на сцену выйдет новый партнер.\n\nВсего шансов: %d", company.name, u.chances)
			sendMsg(bot, chatID, text)

		} else {
			text := fmt.Sprintf("Ты уже участвуешь в сегодняшнем розыгрыше ✅\n\nРезультаты сообщим после подведения итогов.")
			sendMsg(bot, chatID, text)
		}
		


	} else {
		text := fmt.Sprintf("Пока не вижу подписку 👀\nПодпишись на аккаунт %s, вернись в бот и попробуй еще раз.", company.name)
		sendMsgKeyboard(bot, chatID, text, makeSubButton, checkSubButton)
	}
}


func handleTakePart(bot *tgbotapi.BotAPI, chatID int64) {

	text := fmt.Sprintf("Чтобы попасть в розыгрыш призов от %s, подпишись на аккаунт партнера.", company.name)
	sendMsgKeyboard(bot, chatID, text, makeSubButton, checkSubButton)

}

func handleGetChances(bot *tgbotapi.BotAPI, chatID int64, userID int64) {
	u := users[userID]

	text := fmt.Sprintf("Твой фестивальный пропуск\nНакоплено шансов: %d из 10\nЧем больше дней розыгрыша ты посетишь, тем выше вероятность выигратьт главный приз.", u.chances)
	sendMsg(bot, chatID, text)
}



func handleGetInstruction(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Инструкция по участию:\n\n" +
		"1. Получить промокод партнёра, если он доступен.\n" +
		"2. Зарегистрироваться в розыгрыше призов дня.\n" +
		"3. Откроть раздел «Мои шансы», чтобы посмотреть свой статус.\n\n" +
		"Получение промокода или регистрация в розыгрыше позволяют получить не более одного шанса за день."

	sendMsg(bot, chatID, text)
}

func handleGetRules(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Правила акции: \n\n" +
		"• Получение промокода не гарантирует выигрыш.\n" +
		"• Один пользователь может получить не более одного шанса за активный день.\n" +
		"• Активный день считается при получении промокода или регистрации в ежедневном розыгрыше" +
		"• Промокоды доступны, пока не закончатся.\n" +
		"• Результаты розыгрыша объявляются организаторами."

	sendMsg(bot, chatID, text)
}

func handleGetPrize(bot *tgbotapi.BotAPI, chatID int64, userID int64){
	getFirstName(bot, chatID, userID)
	/*getLastName(bot, chatID, user.ID)
	getConsent(bot, chatID, user.ID)*/
}

func getFirstName(bot *tgbotapi.BotAPI, chatID int64, userID int64){

	u := users[userID]
	if u == nil {
		return
	}

	u.winnerCreds = &winnerCreds{}
	text := "Для получения приза необходимо прислать ваше имя, фамилию и согласие на передачу данных партнеру.\n\nВведите ваше имя: \n"
	sendMsg(bot, chatID, text)
}

func inputFirstName(bot *tgbotapi.BotAPI, chatID int64, userID int64, firstName string){

	stateMutex.Lock()
	defer stateMutex.Unlock()

	u := users[userID]
	
	u.winnerCreds.firstName = firstName

	sendMsg(bot, chatID, fmt.Sprintf("Вы ввели имя %s", u.winnerCreds.firstName ))
	time.Sleep(1 * time.Second)
	sendMsg(bot, chatID, "Введите вашу фамилию:\n")

}

func inputLastName(bot *tgbotapi.BotAPI, chatID int64, userID int64, lastName string){

	stateMutex.Lock()
	defer stateMutex.Unlock()

	u := users[userID]
	
	u.winnerCreds.lastName = lastName

	sendMsg(bot, chatID, fmt.Sprintf("Вы ввели фамилию %s", u.winnerCreds.lastName))

	time.Sleep(1 * time.Second)
	text := "Для дальнейшей обработки необходимо предоставить согласие на передачу данных партнеру.\n\n" +
	"Для подтверждения нажмите ниже на кнопку: "
	
	sendMsgKeyboard(bot, chatID, text, consentYesButton, consentNoButton)

}

func inputConsent(bot *tgbotapi.BotAPI, chatID int64, userID int64){

	stateMutex.Lock()
	defer stateMutex.Unlock()

	u := users[userID]
	if u != nil {

		u.IsPrizeClaimed = true 
		fmt.Printf("Пользователь %d успешно забрал приз\n", userID)
	}
	
	text := "Ваши данные приняты. Партнер свяжется с вами для передачи приза. 🎁"
	sendMsg(bot, chatID, text)
}

func handleInputCreds(bot *tgbotapi.BotAPI, chatID int64, userID int64, text string) {
	u := users[userID]
	if u == nil || u.winnerCreds == nil {
		return
	}
	if u.winnerCreds.firstName == "" {
		inputFirstName(bot, chatID, userID, text)
	} else if u.winnerCreds.lastName == "" {
		inputLastName(bot, chatID, userID, text)
	} else {
		inputConsent(bot, chatID, userID)
	}
	
}


