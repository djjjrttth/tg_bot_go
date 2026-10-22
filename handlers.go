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

	}
}


func handleGetPromo(bot *tgbotapi.BotAPI, chatID int64, user *tgbotapi.User) {

	//TODO: добавить ограничение на один промокод по дате
	if len(available_promos) > 0 {

		u := users[user.ID]
		if u.promocode != nil {
			text := fmt.Sprintf("Вы сегодня уже получили промокод\n\nВаш промокод от %s: %v", u.promocode.company, u.promocode.value)
			sendMsg(bot, chatID, text)
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

		text := fmt.Sprintf("Твой фестивальный бонус от %s готов 🎁\n" + 
			"Промокод: %v\nИспользуй его до %s на %s.\n\n" +
			"Теперь можешь принять участие в розыгрыше призов дня.", promo.company, promo.value, promo.expires_at, promo.usage_place)
		
		sendMsg(bot, chatID, text)
		
		
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
	sendMsg(bot, chatID, text)
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
			sendMsg(bot, chatID, text)

		} else {
			text := fmt.Sprintf("Ты уже участвуешь в сегодняшнем розыгрыше ✅\n\nРезультаты сообщим после подведения итогов.")
			sendMsg(bot, chatID, text)
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