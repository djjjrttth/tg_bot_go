package main

import (
	"fmt"
	"log"
	"time"
	"errors"
	"strings"
	"gorm.io/gorm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


//handleCallback обрабатывает входящие команды
func handleCallback(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, adminID int64) {
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
			confirmAndSendDailyDraw(bot, userID)
		
	//case "previewFinalDraw":
	//	previewFinalDraw(bot, userID, company)
		
		// TODO: Добавить проверку user.ID == AdminID
	case "confirmFinalDraw":
		
		confirmAndSendFinalDraw(bot, userID)
	case "consentYes":
		handleConsentResponse(bot, chatID, userID, true)
	case "consentNo":
		handleConsentResponse(bot, chatID, userID, false)
		
	}
}

// handleGetPromo выдает промокод пользователю
func handleGetPromo(bot *tgbotapi.BotAPI, chatID int64, userID int64) {
	
	DB.Model(&Partner{}).Where("id = ?", company.ID).
		Update("promo_request_count", gorm.Expr("promo_request_count + 1"))

	var previousPromo Promo

	today := time.Now().Format("2006-01-02")
	// Поиск промокода по дню выдачи и пользователю, если нашли - сообщаем о выдаче
	err := DB.Where("issued_to = ? AND issued_day = ?", userID, today).First(&previousPromo).Error
	if err == nil {
		text := fmt.Sprintf("Вы сегодня уже получили промокод\n\nВаш промокод от %s: %s", previousPromo.Company, previousPromo.Value)
		sendMsg(bot, chatID, text)
		return
	}

	// Если пользователь не получал промокод
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


	// Если промокоды закончились
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

	// Если промокоды доступны, то выдаем первый свободный, добавляем шанс и обновляем последний активный день
	DB.Model(&User{}).Where("tg_id = ? AND last_active_day != ?", userID, today).
		Updates(map[string]interface{}{
			"chances":         gorm.Expr("LEAST(chances + 1, 10)"),
			"last_active_day": today,
		})

	
	text := fmt.Sprintf("Твой фестивальный бонус от %s готов 🎁\n"+
	"Промокод: %v\nИспользуй его до %s на %s.\n\n"+
	"Теперь можешь принять участие в розыгрыше призов дня.", promo.Company, promo.Value, promo.ExpiresAt.Format("2006-01-02"), promo.UsagePlace)
	sendMsgKeyboard(bot, chatID, text, takePartButton)
	
	
}


// handleMakeSub подписывает пользователя на партнера
func handleMakeSub(bot *tgbotapi.BotAPI, chatID int64, userID int64) {

	if err := markSubscribed(userID, company.ID); err != nil {
		log.Printf("Ошибка создания подписки: %v", err)
		sendMsg(bot, chatID, "Произошла техническая ошибка. Попробуйте позже.")
		return
	}

	text := fmt.Sprintf("Вы подписались на %s\n", company.Name)
	sendMsgKeyboard(bot, chatID, text, checkSubButton)
}


// handleCheckSub проверяет полписку пользователя
func handleCheckSub(bot *tgbotapi.BotAPI, chatID int64, userID int64) {

	fmt.Println(isSubscribed(userID, company.ID))
	// Если пользователь не подписан
	if !isSubscribed(userID, company.ID) {
		text := fmt.Sprintf(
			"Пока не вижу подписку 👀\nПодпишись на аккаунт %s, вернись в бот и попробуй еще раз.",
			company.Name,
		)
		sendMsgKeyboard(bot, chatID, text, makeSubButton, checkSubButton)
		return
	}

	today := time.Now().Format("2006-01-02")


	// Если пользователь уже был подписан на партнера
	var existingPart Participant
	err := DB.Where("tg_id = ? AND partner_id = ? AND day = ?", userID, company.ID, today).First(&existingPart).Error
	if err == nil {
		text := "Ты уже участвуешь в сегодняшнем розыгрыше ✅\n\nРезультаты сообщим после подведения итогов."
		sendMsg(bot, chatID, text)
		return
	}

	// Если пользователь не был раньше подписан на партнера
	//Добавляем нового участника
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

	// Добавляем участнику шанс и обновляем последний активный день
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


// handleTakePart обрабатывает нажатие на кнопку "Участвовать в розыгрыше"
func handleTakePart(bot *tgbotapi.BotAPI, chatID int64) {
	text := fmt.Sprintf("Чтобы попасть в розыгрыш призов от %s, подпишись на аккаунт партнера.", company.Name)
	sendMsgKeyboard(bot, chatID, text, makeSubButton, checkSubButton)
}


// handleGetChances отвечает за раздел "Мои шансы"
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

	text := fmt.Sprintf("Твой фестивальный пропуск\n\n"+
		"Накоплено шансов: %d из 10\n"+
		"Пройдено дней: %d\n"+
		"Осталось дней: %d\n\n"+
		"Партнеры, с которыми ты взаимодействовал:\n%s\n\n"+
		"Главный приз: Билет на Канье Уэста!\n"+
		"Чем больше дней ты посетишь, тем выше вероятность забрать главный приз.",
		u.Chances, passedDays, leftDays, partnersList)

	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
		sendMsg(bot, chatID, text)
	}
}


// handleGetInstruction отвечает за раздел "Как участвовать"
func handleGetInstruction(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Инструкция по участию:\n\n" +
		"1. Получить промокод партнёра, если он доступен.\n" +
		"2. Зарегистрироваться в розыгрыше призов дня.\n" +
		"3. Откроть раздел «Мои шансы», чтобы посмотреть свой статус.\n\n" +
		"Получение промокода или регистрация в розыгрыше позволяют получить не более одного шанса за день."
	sendMsg(bot, chatID, text)
}


// handleGetRules отвечает за раздел "Правила акции"
func handleGetRules(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Правила акции: \n\n" +
		"• Получение промокода не гарантирует выигрыш.\n" +
		"• Один пользователь может получить не более одного шанса за активный день.\n" +
		"• Активный день считается при получении промокода или регистрации в ежедневном розыгрыше" +
		"• Промокоды доступны, пока не закончатся.\n" +
		"• Результаты розыгрыша объявляются организаторами."
	sendMsg(bot, chatID, text)
}

// handleGetPrize управляет состояниями получения приза 
func handleGetPrize(bot *tgbotapi.BotAPI, chatID int64, userID int64) {
	var u User
	if err := DB.First(&u, "tg_id = ?", userID).Error; err != nil {
		sendMsg(bot, chatID, "Ошибка: пользователь не найден в базе данных.")
		return
	}

	// Состояние неподтверждения пользователем получения приза 
	if u.State == "disqualified" {
		sendMsg(bot, chatID, "К сожалению, срок получения приза истек, и он был передан резервному победителю.")
		return
	}

	// Состояние подтверждения пользователем получения приза 
	if u.IsPrizeClaimed || u.State == "completed" {
		sendMsg(bot, chatID, "Вы уже успешно завершили процесс получения приза. Ожидайте связи с вами! ")
		return
	}

	// Состояние ввода фамилии 
	if u.State == "awaiting_last_name" {
		msg := fmt.Sprintf("Вы ввели имя: %s\n\n Теперь введите вашу фамилию:", u.WinnerCreds.FirstName)
		sendMsg(bot, chatID, msg)
		return
	}

	// Состояние подтверждение согласия
	if u.State == "awaiting_consent" {
		msg := "Для дальнейшей обработки необходимо предоставить согласие на передачу данных партнеру.\n\n*Для подтверждения нажмите кнопку ниже:*"
		sendMsgKeyboard(bot, chatID, msg, consentYesButton, consentNoButton)
		return
	}

	if u.WinnerCreds == nil {
		u.WinnerCreds = &WinnerCreds{}
	}

	// Состояние ввода имени
	u.State = "awaiting_first_name"
	
	if err := DB.Save(&u).Error; err != nil {
		log.Printf("Ошибка сохранения состояния пользователя: %v", err)
		sendMsg(bot, chatID, "Произошла техническая ошибка. Попробуйте позже.")
		return
	}

	
	text := "Для получения приза необходимо прислать ваше имя, фамилию и согласие на передачу данных партнеру.\n\n Введите ваше имя:"
	sendMsg(bot, chatID, text)
}

// handleInputCreds управляет вводом данных
func handleInputCreds(bot *tgbotapi.BotAPI, chatID int64, userID int64, text string) {
	var u User
	if err := DB.First(&u, "tg_id = ?", userID).Error; err != nil {
		return
	}

	text = strings.TrimSpace(text)
	if text == "" {
		sendMsg(bot, chatID, "Пожалуйста, введите корректное значение, не оставляйте поле пустым.")
		return
	}

	switch u.State {
	case "awaiting_first_name":
		u.WinnerCreds.FirstName = text
		u.State = "awaiting_last_name"
		DB.Save(&u)
		
		msg := fmt.Sprintf("Вы ввели имя: %s\n\n Теперь введите вашу фамилию:", text)
		sendMsg(bot, chatID, msg)

	case "awaiting_last_name":
		u.WinnerCreds.LastName = text
		u.State = "awaiting_consent"
		DB.Save(&u)
		
		msg := fmt.Sprintf("Вы ввели фамилию: %s\n\n Для дальнейшей обработки необходимо предоставить согласие на передачу данных партнеру.\n\nДля подтверждения нажмите кнопку ниже:", text)
		sendMsgKeyboard(bot, chatID, msg, consentYesButton, consentNoButton)

	case "awaiting_consent":
		sendMsg(bot, chatID, "Пожалуйста, нажмите одну из кнопок согласия ниже, чтобы продолжить.")
		
	case "completed":
		sendMsg(bot, chatID, "Вы уже успешно завершили процесс получения приза. Ожидайте связи с вами!")

	default:
		u.State = ""
		DB.Save(&u)
	}
}

// handleConsentResponse управляет согласием передачи контактных данных партнеру
func handleConsentResponse(bot *tgbotapi.BotAPI, chatID int64, userID int64, accepted bool) {
	var u User
	if err := DB.First(&u, "tg_id = ?", userID).Error; err != nil {
		return
	}

	if u.State != "awaiting_consent" {
		sendMsg(bot, chatID, "Этот шаг уже пройден или неактуален.")
		return
	}

	if !accepted {
		u.State = ""
		u.WinnerCreds.Consent = boolPtr(false)
		DB.Save(&u)
		sendMsg(bot, chatID, "Вы отказались от передачи данных. К сожалению, мы не можем передать вам приз.")
		return
	}

	consentTrue := true
	u.WinnerCreds.Consent = &consentTrue
	u.IsPrizeClaimed = true
	u.State = "completed" 
	
	if err := DB.Save(&u).Error; err != nil {
		log.Printf("Ошибка сохранения данных победителя: %v", err)
		sendMsg(bot, chatID, "Произошла техническая ошибка при сохранении данных.")
		return
	}

	log.Printf("Пользователь %d успешно подтвердил получение приза", userID)
	sendMsg(bot, chatID, "Ваши данные приняты! Партнер свяжется с вами для передачи приза в ближайшее время.")
}

