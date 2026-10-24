package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func convertToRow(button tgbotapi.InlineKeyboardButton) []tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardRow(button)
}

func convertToMarkup(buttons []tgbotapi.InlineKeyboardButton) tgbotapi.InlineKeyboardMarkup {
	k := [][]tgbotapi.InlineKeyboardButton{}

	for _, b := range buttons {
		k = append(k, convertToRow(b))
	}

	return tgbotapi.NewInlineKeyboardMarkup(k...)

}

var getPromoButton = tgbotapi.NewInlineKeyboardButtonData("Получить промокод", "getPromo")
var takePartButton = tgbotapi.NewInlineKeyboardButtonData("Участвовать в розыгрыше", "takePart")
var getInstructionButton = tgbotapi.NewInlineKeyboardButtonData("Как участвовать", "getInstruction")
var getChancesButton = tgbotapi.NewInlineKeyboardButtonData("Мои шансы", "getChances")
var getRulesButton = tgbotapi.NewInlineKeyboardButtonData("Правила акции", "getRules")
var makeSubButton = tgbotapi.NewInlineKeyboardButtonData("Подписаться", "makeSub")
var checkSubButton = tgbotapi.NewInlineKeyboardButtonData("Проверить подписку", "checkSub")
var getPrizeButton = tgbotapi.NewInlineKeyboardButtonData("Получить приз", "getPrize")
var consentYesButton = tgbotapi.NewInlineKeyboardButtonData("Подтвердить", "consentYes")
var consentNoButton = tgbotapi.NewInlineKeyboardButtonData("Не подтверждать", "consentNo")


var buttons = tgbotapi.NewInlineKeyboardMarkup(
	
	convertToRow(getPromoButton),

	convertToRow(takePartButton),

	convertToRow(getInstructionButton),

	convertToRow(getChancesButton),
	
	convertToRow(getRulesButton),

	tgbotapi.NewInlineKeyboardRow(
	tgbotapi.NewInlineKeyboardButtonData("GetDailyWinners", "getDailyWinners"),
),

tgbotapi.NewInlineKeyboardRow(
	tgbotapi.NewInlineKeyboardButtonData("runDailyDraw", "runDailyDraw"),
),
tgbotapi.NewInlineKeyboardRow(
	tgbotapi.NewInlineKeyboardButtonData("runFinalDraw", "runFinalDraw"),
),
	
)


/**/