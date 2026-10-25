package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


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
var confirmDailyDrawButton = tgbotapi.NewInlineKeyboardButtonData("Подтвердить и уведомить участников", "confirmDailyDraw")
var confirmFinalDrawButton = tgbotapi.NewInlineKeyboardButtonData("Подтвердить и уведомить участников", "confirmFinalDraw")


var buttons = tgbotapi.NewInlineKeyboardMarkup(
	
	convertToRow(getPromoButton),

	convertToRow(takePartButton),

	convertToRow(getInstructionButton),

	convertToRow(getChancesButton),
	
	convertToRow(getRulesButton),

)


