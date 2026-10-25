package main

import (
	"fmt"
	"strings"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


// getTotalUsers получает общее число пользователей
func getTotalUsers() int64{
	var totalUsers int64
	DB.Model(&User{}).Count(&totalUsers)

	return totalUsers
}


// getNewUsersDaily считает количество новых пользователей по дням
func getNewUsersDaily() []DailyStat {
	var newUsersDaily []DailyStat
	DB.Model(&User{}).Select("DATE(joined_at) as day, COUNT(*) as count").Group("day").Order("day DESC").Limit(7).Scan(&newUsersDaily)
	return newUsersDaily
}

// countPromoRequest считает количество запросов промокода;
func countPromoRequest() int64 {
	var promoRequests int64	
	DB.Model(&Partner{}).Select("COALESCE(SUM(promo_request_count), 0)").Scan(&promoRequests)

	return promoRequests
}

// countPromoIssued считает количество выданных промокодов
func countPromoIssued() int64 {
	var promoIssued int64
	DB.Model(&Promo{}).Count(&promoIssued)

	return promoIssued
}

// getPromoRemaining считает остаток промокодов по каждому партнеру
func getPromoRemaining() []PromoRemaining {
	var remaining []PromoRemaining
	DB.Model(&Promo{}).Select("partner_id, COUNT(*) as count").Where("issued_at IS NULL").Group("partner_id").Scan(&remaining)

	return remaining
}


// getPartStat считает количество участников каждого ежедневного розыгрыша
func getPartStat() []PartStat{
	var partsDaily []PartStat
	DB.Model(&Participant{}).Select("day, partner_id, COUNT(DISTINCT tg_id) as count").Group("day, partner_id").Order("day DESC").Scan(&partsDaily)
	
	return partsDaily
}


// handleStats считает статистику и формирует сообщение со статистикой
func handleStats(bot *tgbotapi.BotAPI, chatID int64) {
	var sb strings.Builder
	sb.WriteString("Статистика\n\n")

	totalUsers := getTotalUsers()
	sb.WriteString(fmt.Sprintf("Всего пользователей: %d\n\n", totalUsers))

	newUsersDaily := getNewUsersDaily()
	sb.WriteString("Новые пользователи по дням:\n")
	
	for _, stat := range newUsersDaily {
		sb.WriteString(fmt.Sprintf("- %s: %d чел.\n", stat.Day, stat.Count))
	}
	
	sb.WriteString("\n")

	promoRequests := countPromoRequest()
	sb.WriteString(fmt.Sprintf("Количество запросов промокода: %d\n\n", promoRequests))

	promoIssued := countPromoIssued()
	sb.WriteString(fmt.Sprintf("Количество выданных промокодов: %d\n\n", promoIssued))


	promoRemaining := getPromoRemaining()
	
	sb.WriteString("Остатки промокодов по партнерам:\n")
	
	for _, r := range promoRemaining {
		var p Partner
		DB.First(&p, r.PartnerID)
		sb.WriteString(fmt.Sprintf("  - %s: %d шт.\n", p.Name, r.Count))
	}
	
	sb.WriteString("\n")
	
	partsDaily := getPartStat()
		
	sb.WriteString("Участники ежедневных розыгрышей:\n")
	
	partnerNames := make(map[uint]string)
	for _, ps := range partsDaily {
		if _, ok := partnerNames[ps.PartnerID]; !ok {
			var p Partner
			if DB.First(&p, ps.PartnerID).Error == nil {
				partnerNames[ps.PartnerID] = p.Name
			} else {
				partnerNames[ps.PartnerID] = "Неизвестный"
			}
		}
	}
	
	for _, ps := range partsDaily {
		sb.WriteString(fmt.Sprintf("- %s (%s): %d чел.\n", ps.Day, partnerNames[ps.PartnerID], ps.Count))
	}
	

	sendMsg(bot, chatID, sb.String())
}

