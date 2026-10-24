package main

import (
	"fmt"
	"log"
	//"math/rand/v2"
	//"slices"
	"time"
	"errors"
	"gorm.io/gorm"
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
	
	log.Printf("Выдаем промокод")
	var previousPromo Promo

	today := time.Now().Format("2006-01-02")
	err := DB.Where("issued_to = ? AND issued_day = ?", userID, today).First(&previousPromo).Error
	if err == nil {
		text := fmt.Sprintf("Вы сегодня уже получили промокод\n\nВаш промокод от %s: %s", previousPromo.Company, previousPromo.Value)
		sendMsg(bot, chatID, text)
		return
	}

	var promo Promo
	err = DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Raw(`
			SELECT * FROM promos 
			WHERE partner_id = ? AND issued_at IS NULL 
			ORDER BY id 
			FOR UPDATE SKIP LOCKED 
			LIMIT 1`, company.ID).Scan(&promo).Error
		
		if err != nil {
			return err
		}
		
		if promo.ID == 0 {
			return gorm.ErrRecordNotFound
		}

		now := time.Now()
		promo.IssuedAt = &now
		promo.IssuedTo = &userID
		promo.IssuedDay = today
		return tx.Save(&promo).Error
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			text := fmt.Sprintf("Промокоды от %s уже закончились — их разобрали быстрее всех 🔥\nНо ты все еще можешь подписаться на партнера и принять участие в розыгрыше призов дня.", company.Name)
			sendMsgKeyboard(bot, chatID, text, takePartButton)
			return
		}
		log.Printf("Ошибка БД при выдаче промокода: %v", err)
		sendMsg(bot, chatID, "Произошла техническая ошибка. Попробуйте позже.")
		return
	}

	DB.Model(&User{}).Where("tg_id = ? AND last_active_day != ?", userID, today).
		Updates(map[string]interface{}{
			"chances":         gorm.Expr("LEAST(chances + 1, 10)"),
			"last_active_day": today,
		})

	
	text := fmt.Sprintf("Твой фестивальный бонус от %s готов 🎁\n"+
	"Промокод: %v\nИспользуй его до %s на %s.\n\n"+
	"Теперь можешь принять участие в розыгрыше призов дня.", promo.Company, promo.Value, promo.ExpiresAt, promo.UsagePlace)
	sendMsgKeyboard(bot, chatID, text, takePartButton)
	
	
}

func handleMakeSub(bot *tgbotapi.BotAPI, chatID int64, userID int64) {

	if err := markSubscribed(userID, company.ID); err != nil {
		log.Printf("Ошибка создания подписки: %v", err)
		sendMsg(bot, chatID, "Произошла техническая ошибка. Попробуйте позже.")
		return
	}

	text := fmt.Sprintf("Вы подписались на %s\n", company.Name)
	sendMsgKeyboard(bot, chatID, text, checkSubButton)
}



func handleCheckSub(bot *tgbotapi.BotAPI, chatID int64, userID int64) {

	fmt.Println(isSubscribed(userID, company.ID))
	if !isSubscribed(userID, company.ID) {
		text := fmt.Sprintf(
			"Пока не вижу подписку 👀\nПодпишись на аккаунт %s, вернись в бот и попробуй еще раз.",
			company.Name,
		)
		sendMsgKeyboard(bot, chatID, text, makeSubButton, checkSubButton)
		return
	}

	today := time.Now().Format("2006-01-02")

	var existingPart Participant
	err := DB.Where("tg_id = ? AND partner_id = ? AND day = ?", userID, company.ID, today).First(&existingPart).Error
	if err == nil {
		text := "Ты уже участвуешь в сегодняшнем розыгрыше ✅\n\nРезультаты сообщим после подведения итогов."
		sendMsg(bot, chatID, text)
		return
	}

	newPart := Participant{
		TgID:      userID,
		PartnerID: company.ID,
		Day:       today,
		Timestamp: time.Now(),
	}

	if err := DB.Create(&newPart).Error; err != nil {
		log.Printf("Ошибка регистрации участника: %v", err)
		sendMsg(bot, chatID, "Произошла ошибка при регистрации. Попробуйте позже.")
		return
	}

	
	DB.Model(&User{}).Where("tg_id = ? AND last_active_day != ?", userID, today).
		Updates(map[string]interface{}{
			"chances":         gorm.Expr("LEAST(chances + 1, 10)"),
			"last_active_day": today,
		})

	var u User
	DB.First(&u, "tg_id = ?", userID)

	text := fmt.Sprintf(
		"Ты в лайнапе ✅\n\nЗаявка на розыгрыш призов от %s зарегистрирована.\n\nВозвращайся завтра — на сцену выйдет новый партнер.\n\nВсего шансов: %d",
		company.Name, u.Chances,
	)
	sendMsg(bot, chatID, text)
}

func handleTakePart(bot *tgbotapi.BotAPI, chatID int64) {
	text := fmt.Sprintf("Чтобы попасть в розыгрыш призов от %s, подпишись на аккаунт партнера.", company.Name)
	sendMsgKeyboard(bot, chatID, text, makeSubButton, checkSubButton)
}

func handleGetChances(bot *tgbotapi.BotAPI, chatID int64, userID int64) {
	var u User
	if err := DB.First(&u, "tg_id = ?", userID).Error; err != nil {
		sendMsg(bot, chatID, "Вы еще не участвовали в розыгрышах.")
		return
	}

	var passedDays int64
	DB.Model(&Participant{}).
		Where("tg_id = ?", userID).
		Distinct("day").
		Count(&passedDays)

	var partnerIDs []uint
	DB.Model(&Participant{}).
		Where("tg_id = ?", userID).
		Distinct("partner_id").
		Pluck("partner_id", &partnerIDs)

	var partnerNames []string
	for _, pid := range partnerIDs {
		var partner Partner
		if err := DB.First(&partner, "id = ?", pid).Error; err == nil {
			partnerNames = append(partnerNames, partner.Name)
		}
	}

	partnersList := "—"
	if len(partnerNames) > 0 {
		partnersList = ""
		for _, name := range partnerNames {
			partnersList += fmt.Sprintf("• %s\n", name)
		}
	}

	leftDays := int64(10) - passedDays
	if leftDays < 0 {
		leftDays = 0
	}

	text := fmt.Sprintf("*Твой фестивальный пропуск*\n\n"+
		"Накоплено шансов: *%d из 10*\n"+
		"Пройдено дней: *%d*\n"+
		"Осталось дней: *%d*\n\n"+
		"Партнеры, с которыми ты взаимодействовал:\n%s\n"+
		"*Главный приз:* Билет на Канье Уэста!\n"+
		"Чем больше дней ты посетишь, тем выше вероятность забрать главный приз.",
		u.Chances, passedDays, leftDays, partnersList)

	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
		sendMsg(bot, chatID, text)
	}
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

func handleGetPrize(bot *tgbotapi.BotAPI, chatID int64, userID int64) {
	getFirstName(bot, chatID, userID)
	// /getLastName(bot, chatID, user.ID)
	// getConsent(bot, chatID, user.ID)/
}

func getFirstName(bot *tgbotapi.BotAPI, chatID int64, userID int64) {
	u := users[userID]
	if u == nil {
		return
	}
	u.WinnerCreds = &WinnerCreds{}
	text := "Для получения приза необходимо прислать ваше имя, фамилию и согласие на передачу данных партнеру.\n\nВведите ваше имя: \n"
	sendMsg(bot, chatID, text)
}

func inputFirstName(bot *tgbotapi.BotAPI, chatID int64, userID int64, firstName string) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	
	u := users[userID]
	if u == nil || u.WinnerCreds == nil {
		return
	}
	u.WinnerCreds.FirstName = firstName
	sendMsg(bot, chatID, fmt.Sprintf("Вы ввели имя %s", u.WinnerCreds.FirstName))
	time.Sleep(1 * time.Second)
	sendMsg(bot, chatID, "Введите вашу фамилию:\n")
}

func inputLastName(bot *tgbotapi.BotAPI, chatID int64, userID int64, lastName string) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	
	u := users[userID]
	if u == nil || u.WinnerCreds == nil {
		return
	}
	u.WinnerCreds.LastName = lastName
	sendMsg(bot, chatID, fmt.Sprintf("Вы ввели фамилию %s", u.WinnerCreds.LastName))
	time.Sleep(1 * time.Second)
	text := "Для дальнейшей обработки необходимо предоставить согласие на передачу данных партнеру.\n\n" +
		"Для подтверждения нажмите ниже на кнопку: "
	sendMsgKeyboard(bot, chatID, text, consentYesButton, consentNoButton)
}

func inputConsent(bot *tgbotapi.BotAPI, chatID int64, userID int64) {
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
	if u == nil || u.WinnerCreds == nil {
		return
	}
	if u.WinnerCreds.FirstName == "" {
		inputFirstName(bot, chatID, userID, text)
	} else if u.WinnerCreds.LastName == "" {
		inputLastName(bot, chatID, userID, text)
	} else {
		inputConsent(bot, chatID, userID)
	}
}