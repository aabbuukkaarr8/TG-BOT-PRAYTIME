package bot_working

import (
	"encoding/json"
	"fmt"
	"gihub.com/aabbuukkaarr8/TG-BOT-PRAYTIME/internal/service"
	"gihub.com/aabbuukkaarr8/TG-BOT-PRAYTIME/parse_api"
	rabbitmq "gihub.com/aabbuukkaarr8/TG-BOT-PRAYTIME/rabbit_mq"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

func StartBot(srv *service.Service, rabbit *rabbitmq.Client) {
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
		if update.Message.Text == "Напоминать" {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			//Проверка запущен ли уже бот у этого пользователя
			answer := srv.Exists(update.Message.Chat.ID)
			if answer == true {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Бот уже запущен!"))
				continue
			}
			user := service.User{
				ChatID:  update.Message.Chat.ID,
				Status:  "on",
				Fajr:    "on",
				Zuhr:    "on",
				Asr:     "on",
				Maghrib: "on",
				Isha:    "on",
			}
			srv.Create(user)

			go parse_api.SchedulePrayerNotifications(rabbit, update.Message.Chat.ID)

			btn := tgbotapi.NewInlineKeyboardButtonData("Настройки", "settings")
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(btn),
			)

			// создаём сообщение с кнопкой сразу
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Теперь я буду напоминать тебе о Намазах!")
			msg.ReplyMarkup = keyboard

			bot.Send(msg)
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
