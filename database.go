package main

import (
	"log"
	"os"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB инициализирует базу данных
func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("Ошибка: переменная окружения DATABASE_URL не найдена в .env")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), 
	})
	if err != nil {
		log.Fatal("Ошибка подключения к БД: ", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Ошибка получения sql.DB: ", err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	err = db.AutoMigrate(
		&User{}, &Partner{}, &Promo{}, &Participant{}, &Draw{}, &Subscription{},
	)
	if err != nil {
		log.Fatal("Ошибка миграции БД: ", err)
	}

	DB = db
	log.Println("База данных подключена")
}