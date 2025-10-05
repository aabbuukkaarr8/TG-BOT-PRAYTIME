package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type PrayerNotification struct {
	ChatID      int64     `json:"chat_id"`
	PrayerName  string    `json:"prayer_name"`
	PrayerTime  string    `json:"prayer_time"`
	ScheduledAt time.Time `json:"scheduled_at"`
}

func New(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Создаем exchange
	err = channel.ExchangeDeclare(
		"prayer_notifications",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Создаем очередь
	_, err = channel.QueueDeclare(
		"prayer_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Привязываем очередь к exchange
	err = channel.QueueBind(
		"prayer_queue",
		"prayer",
		"prayer_notifications",
		false,
		nil,
	)

	log.Println("✅ RabbitMQ connected for prayer notifications")
	return &Client{conn: conn, channel: channel}, nil
}

func (c *Client) PublishPrayerNotification(ctx context.Context, notification *PrayerNotification) error {
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	err = c.channel.PublishWithContext(
		ctx,
		"prayer_notifications",
		"prayer",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("📨 Prayer notification published: %s at %s", notification.PrayerName, notification.ScheduledAt.Format("15:04"))
	return nil
}

func (c *Client) ConsumePrayerNotifications() (<-chan amqp.Delivery, error) {
	messages, err := c.channel.Consume(
		"prayer_queue",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume messages: %w", err)
	}

	log.Println("👂 Started consuming prayer notifications")
	return messages, nil
}

func (c *Client) Close() {
	c.channel.Close()
	c.conn.Close()
}
