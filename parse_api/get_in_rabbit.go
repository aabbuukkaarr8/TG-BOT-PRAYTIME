package parse_api

import (
	"context"
	rabbitmq "gihub.com/aabbuukkaarr8/TG-BOT-PRAYTIME/rabbitMQ"
	"time"
)

func SchedulePrayerNotifications(rabbit *rabbitmq.Client, chatID int64) {
	taskForSchedule(rabbit, chatID)
	for {
		now := time.Now()
		next := now.Add(24 * time.Hour)
		next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
		duration := next.Sub(now)
		time.Sleep(duration)
		taskForSchedule(rabbit, chatID)
	}
}
func taskForSchedule(rabbit *rabbitmq.Client, chatID int64) {
	times, _ := PrayerTimes()
	now := time.Now()

	for name, timeStr := range times {
		prayerTime := parsePrayerTime(timeStr)

		// ДОБАВИЛИ ПРОВЕРКИ:
		if prayerTime.Before(now) || prayerTime.Before(now) {
			continue
		}

		notification := &rabbitmq.PrayerNotification{
			ChatID:      chatID,
			PrayerName:  name,
			PrayerTime:  timeStr,
			ScheduledAt: prayerTime,
		}
		if notification.PrayerName == "Isha" {

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
