package main

import (
	"fmt"
	"log"
	"slices"
	"math/rand/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func sendMsg(bot *tgbotapi.BotAPI, chatID int64, text string){
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}


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
		case "getInstruction":
			handleGetInstruction(bot, chatID)
		case "getRules":
			handleGetRules(bot, chatID)
		case "getDailyWinners":
			getDailyWinnersID(company)
		case "runDailyDraw":
			dailyDraw(bot, company)
		case "getPrize":
			handleGetPrize(bot, chatID, user)

		case "runFinalDraw":
			runFinalDraw(bot, company)

	}
}


func handleGetPromo(bot *tgbotapi.BotAPI, chatID int64, user *tgbotapi.User) {

	//TODO: добавить ограничение на один промокод по дате
	u := users[user.ID]
	if u.promocode != nil {
		text := fmt.Sprintf("Вы сегодня уже получили промокод\n\nВаш промокод от %s: %v", u.promocode.company, u.promocode.value)
		sendMsg(bot, chatID, text)

	} else {
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
			"Теперь можешь принять участие в розыгрыше призов дня.", promo.company, promo.value, promo.expires_at, promo.usage_place)
		
		sendMsg(bot, chatID, text)
		
		
		} else {
			text := fmt.Sprintf("Промокоды от %s уже закончились — их разобрали быстрее всех 🔥\nНо ты все еще можешь подписаться на партнера и принять участие в розыгрыше призов дня.", company.name)
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
		
}


func handleMakeSub(bot *tgbotapi.BotAPI, chatID int64, user *tgbotapi.User){
	//fmt.Println(users[user.ID])
	u := users[user.ID]
	//fmt.Println(u.registered)

	u.registered = true
	//fmt.Printf("Статус подписки: %v\n", statusSub)

	text := fmt.Sprintf("Вы подписались на %s\n", company.name)
	sendMsg(bot, chatID, text)
}

func handleCheckSub(bot *tgbotapi.BotAPI, chatID int64, user *tgbotapi.User){
	u := users[user.ID]
	if u.registered {
		if _, ok := findParticipantByID(user.ID); !ok {
			participants = append(participants, &Participant{user.ID, "10.10.2026", "12:00:00"})
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

	text := fmt.Sprintf("Чтобы попасть в розыгрыш призов от %s, подпишись на аккаунт партнера.", company.name)
	msg := tgbotapi.NewMessage(chatID, text)

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Подписаться на %s", company.name), "makeSub"),
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


func getDailyWinnersID(p Partner) []int64 {

	fmt.Println(participants)
	var n uint32 = p.winners_cnt
	var participants_cnt uint32 = uint32(len(participants))
	winnerIDs := make([]int64, 0)
	
	if participants_cnt >= n {
		

		win_inds := rand.Perm(int(participants_cnt))

		fmt.Println(win_inds)
		fmt.Println()
		for i := range n {
			winnerIDs = append(winnerIDs, participants[win_inds[i]].tgID)
			//fmt.Println()
		}
		fmt.Println()

		fmt.Println(winnerIDs)
		return winnerIDs

	}

	return nil
}

func dailyDraw(bot *tgbotapi.BotAPI, p Partner){
	winnerIDs := getDailyWinnersID(p)
	fmt.Println(winnerIDs)

	var newDraw = Draw{
		date : "10.10.2026",
		time : "12:00:00", 
		participants : participants,
		winners : winnerIDs,
	}

	dailyDraws = append(dailyDraws, newDraw)
	sendMsgToDailyWinners(bot, winnerIDs)
	fmt.Println(dailyDraws)
}

func sendMsgToDailyWinners(bot *tgbotapi.BotAPI, dailyWinnerIDs []int64){

	text := fmt.Sprintf("Ты в числе хедлайнеров дня 🔥\n" + 

		"Ты выиграл %s от %s.\n" +

		"Чтобы получить приз, %s.", "айфон 17 про макс", "Yandex", "нажми на кнопку 'Получить приз'")

	for _, id := range dailyWinnerIDs {
		msg := tgbotapi.NewMessage(id, text)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Получить приз", "getPrize"),
			),
		)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
	}
}

func findParticipantByID(id int64) (int, bool){
	var n int = len(participants)

	for i := range n {
		p := participants[i]
		if p.tgID == id {
			return i, true
		}
	}

	return -1, false
}

func handleGetPrize(bot *tgbotapi.BotAPI, chatID int64, user *tgbotapi.User){
	getFirstName(bot, chatID, user)
	/*getLastName(bot, chatID, user.ID)
	getConsent(bot, chatID, user.ID)*/
}

func getFirstName(bot *tgbotapi.BotAPI, chatID int64, user *tgbotapi.User){

	u := users[user.ID]
	if u == nil {
		return
	}

	u.winnerCreds = &winnerCreds{}
	text := "Для получения приза необходимо прислать ваше имя, фамилию и согласие на передачу данных партнеру.\n\nВведите ваше имя: \n"
	sendMsg(bot, chatID, text)
}

func inputFirstName(bot *tgbotapi.BotAPI, chatID int64, user *tgbotapi.User, firstName string){
	u := users[user.ID]
	
	u.winnerCreds.firstName = firstName

	sendMsg(bot, chatID, fmt.Sprintf("Вы ввели имя %s", u.winnerCreds.firstName ))

	sendMsg(bot, chatID, "Введите вашу фамилию:\n")

}

func inputLastName(bot *tgbotapi.BotAPI, chatID int64, user *tgbotapi.User, lastName string){
	u := users[user.ID]
	
	u.winnerCreds.lastName = lastName

	sendMsg(bot, chatID, fmt.Sprintf("Вы ввели фамилию %s", u.winnerCreds.lastName))


	text := "Для дальнейшей обработки необходимо предоставить согласие на передачу данных партнеру.\n\n" +
	"Для подтверждения нажмите ниже на кнопку: "
	
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подтвердить", "consentYes"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Не подтверждать", "consentNo"),
		),
	)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}

}


func inputConsent(bot *tgbotapi.BotAPI, chatID int64, user *tgbotapi.User){
	

	fmt.Println("Согласие получено")
}

func handleInputCreds(bot *tgbotapi.BotAPI, chatID int64, user *tgbotapi.User, text string) {
	u := users[user.ID]
	if u.winnerCreds.firstName == "" {
		inputFirstName(bot, chatID, user, text)
	} else if u.winnerCreds.lastName == "" {
		inputLastName(bot, chatID, user, text)
	} else {
		inputConsent(bot, chatID, user)
	}
	
}

func chooseMainFinalWinners(p Partner, participantIDs []int64) []int64 {
	var finalWinners []int64
	for len(finalWinners) != 5 {
		ind := rand.IntN(len(participantIDs))
		pID:= participantIDs[ind]
		if !slices.Contains(finalWinners, pID) {
			finalWinners = append(finalWinners, pID)
		}
	}

	return finalWinners
}

func chooseReserveFinalWinners(p Partner, participantIDs, mainFinalWinnerIDs []int64) []int64 {
	var reserveWinners []int64
	for len(reserveWinners) != 5 {
		ind := rand.IntN(len(participantIDs))
		pID:= participantIDs[ind]
		if !slices.Contains(mainFinalWinnerIDs, pID) && !slices.Contains(reserveWinners, pID) {
			reserveWinners = append(reserveWinners, pID)
		}
	}

	return reserveWinners
}

func getFinalWinnersID(p Partner) ([]int64, []int64) {
	var participantsWithChancesIDs []int64

	if len(users) >= 10 {
		for _, p := range participants {
			u := users[p.tgID]
			for range u.chances {
				participantsWithChancesIDs = append(participantsWithChancesIDs, p.tgID)
			}

		}

		finalWinnersID := chooseMainFinalWinners(p, participantsWithChancesIDs)
		reserveWinnersID := chooseReserveFinalWinners(p, participantsWithChancesIDs, finalWinnersID)

		fmt.Printf("Главные победители: %v\n", finalWinnersID)
		fmt.Printf("Резервные победители: %v\n\n", reserveWinnersID)

		return finalWinnersID, reserveWinnersID

	}

	log.Printf("Недостаточно участников для проведения финального розыгрыша")
	return nil, nil
}

func runFinalDraw(bot *tgbotapi.BotAPI, p Partner){
	finalWinnerIDs, _ := getFinalWinnersID(p)
	sendMsgToFinalWinners(bot, finalWinnerIDs)

}

func sendMsgToFinalWinners(bot *tgbotapi.BotAPI, finalWinnerIDs []int64){

	text := fmt.Sprintf("Твой сентябрь выходит на главную сцену 🔥\n" + 
		"Ты выиграл билет на Канье Уэста!" + 
		"Чтобы подтвердить получение приза, нажми кнопку ниже и оставь необходимые контактные данные.")

	for _, id := range finalWinnerIDs {
		msg := tgbotapi.NewMessage(id, text)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Получить приз", "getPrize"),
			),
		)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
	}
}