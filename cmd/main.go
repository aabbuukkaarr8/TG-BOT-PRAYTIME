package main

import (
	"encoding/json"
	"fmt"
	"gihub.com/aabbuukkaarr8/TG-BOT-PRAYTIME/parse_api"
	rabbitmq "gihub.com/aabbuukkaarr8/TG-BOT-PRAYTIME/rabbitMQ"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var dailyStarts = make(map[int64]string)

func main() {
	rabbit, err := rabbitmq.New("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer rabbit.Close()

	bot, err := tgbotapi.NewBotAPI("8470796435:AAEHCu9VzGRjJ_DCPthocbF1KgLSicZ2dTE")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	go startPrayerWorker(rabbit, bot)

	for update := range updates {
		if update.Message.Text == "/start" {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			today := time.Now().Format("2006-01-02")

			// ПРОВЕРЯЕМ БЫЛ ЛИ УЖЕ /start СЕГОДНЯ
			if dailyStarts[update.Message.Chat.ID] == today {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Вы уже запускали бота сегодня"))
				continue
			}

			// СОХРАНЯЕМ ДАТУ ЗАПУСКА
			dailyStarts[update.Message.Chat.ID] = today
			go parse_api.SchedulePrayerNotifications(rabbit, update.Message.Chat.ID)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Теперь я буду напоминать тебе о Намазах!"))
		}

		// /times - показать время намазов
		if update.Message.Text == "/times" {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			times, err := parse_api.PrayerTimes()
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка получения времени намазов"))
				continue
			}

			// Форматируем красивое сообщение
			message := "🕌 Время намазов для Ингушетии на сегодня:\n\n"
			message += fmt.Sprintf("🌅 Фаджр: %s\n", times["Fajr"])
			message += fmt.Sprintf("☀️ Восход: %s\n", times["Sunrise"])
			message += fmt.Sprintf("🏙️ Зухр: %s\n", times["Dhuhr"])
			message += fmt.Sprintf("🌇 Аср: %s\n", times["Asr"])
			message += fmt.Sprintf("🌆 Магриб: %s\n", times["Maghrib"])
			message += fmt.Sprintf("🌃 Иша: %s\n", times["Isha"])

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, message))
		}
	}
}

func startPrayerWorker(rabbit *rabbitmq.Client, bot *tgbotapi.BotAPI) {
	messages, _ := rabbit.ConsumePrayerNotifications()

	for msg := range messages {
		var notification rabbitmq.PrayerNotification
		json.Unmarshal(msg.Body, &notification)

		// Проверяем время
		if time.Now().Before(notification.ScheduledAt) {
			msg.Nack(false, true)
			continue
		}

		// Отправляем уведомление
		sendPrayerNotification(notification, bot)
		msg.Ack(false)
	}
}

func sendPrayerNotification(notification rabbitmq.PrayerNotification, bot *tgbotapi.BotAPI) {
	message := fmt.Sprintf("⏰ Пришло время совершить %s намаз! 'Обратитесь за помощью к терпению и намазу' 2:45",
		notification.PrayerName)
	bot.Send(tgbotapi.NewMessage(notification.ChatID, message))
}
