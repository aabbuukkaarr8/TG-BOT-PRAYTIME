package parse_api

import (
	"context"
	rabbitmq "gihub.com/aabbuukkaarr8/TG-BOT-PRAYTIME/rabbitMQ"
	"time"
)

// ЗАМЕНИ существующую функцию на эту:
func SchedulePrayerNotifications(rabbit *rabbitmq.Client, chatID int64) {
	times, _ := PrayerTimes()
	now := time.Now()

	for name, timeStr := range times {
		prayerTime := parsePrayerTime(timeStr)
		notifyTime := prayerTime.Add(-5 * time.Minute)

		// ДОБАВИЛИ ПРОВЕРКИ:
		if prayerTime.Before(now) || notifyTime.Before(now) {
			continue
		}

		notification := &rabbitmq.PrayerNotification{
			ChatID:      chatID,
			PrayerName:  name,
			PrayerTime:  timeStr,
			ScheduledAt: notifyTime,
		}

		rabbit.PublishPrayerNotification(context.Background(), notification)
	}
}
func parsePrayerTime(timeStr string) time.Time {
	// timeStr в формате "05:30" (24h)
	now := time.Now()

	// Парсим время намаза
	prayerTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}
	}

	// Собираем полную дату: сегодня + время намаза
	return time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		prayerTime.Hour(),
		prayerTime.Minute(),
		0, 0,
		now.Location(),
	)
}
