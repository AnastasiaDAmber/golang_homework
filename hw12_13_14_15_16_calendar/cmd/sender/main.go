package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/logger"
	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/rabbitmq"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
}

func main() {
	flag.Parse()

	if configFile == "" {
		fmt.Fprintf(os.Stderr, "error: config file is required\n")
		os.Exit(1)
	}

	cfg, err := NewConfigFromFile(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logg := logger.New(cfg.Logger.Level)

	// Подключаемся к RabbitMQ
	rmqClient := rabbitmq.NewClient(cfg.RabbitMQ.URL)
	if err := rmqClient.Connect(context.Background()); err != nil {
		logg.Error("failed to connect to RabbitMQ: " + err.Error())
		os.Exit(1)
	}
	defer rmqClient.Close()

	logg.Info("sender is running...")

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	// Начинаем чтение сообщений из очереди
	messageChan, err := rmqClient.Consume(ctx, cfg.RabbitMQ.QueueName)
	if err != nil {
		logg.Error("failed to start consuming messages: " + err.Error())
		os.Exit(1)
	}

	for {
		select {
		case <-ctx.Done():
			logg.Info("sender is stopping...")
			return
		case msg, ok := <-messageChan:
			if !ok {
				logg.Info("message channel closed")
				return
			}

			// Десериализуем уведомление
			notification, err := rabbitmq.DeserializeNotification(msg.Body)
			if err != nil {
				logg.Error("failed to deserialize notification: " + err.Error())
				// Отклоняем сообщение без повторной постановки в очередь
				if err := msg.Reject(); err != nil {
					logg.Error("failed to reject message: " + err.Error())
				}
				continue
			}

			// Проверяем валидность
			if err := notification.Validate(); err != nil {
				logg.Error("invalid notification: " + err.Error())
				if err := msg.Reject(); err != nil {
					logg.Error("failed to reject message: " + err.Error())
				}
				continue
			}

			// Отправляем уведомление (логируем в STDOUT)
			sendNotification(logg, notification)

			// Подтверждаем обработку сообщения
			if err := msg.Ack(); err != nil {
				logg.Error("failed to ack message: " + err.Error())
			}
		}
	}
}

func sendNotification(logg *logger.Logger, notification *rabbitmq.Notification) {
	// Выводим уведомление в STDOUT (как требуется в задании)
	message := fmt.Sprintf(
		"[NOTIFICATION] Event: %s | Title: %s | Date: %s | User: %s",
		notification.EventID,
		notification.Title,
		notification.EventDate.Format(time.RFC3339),
		notification.UserID,
	)

	fmt.Println(message)
	logg.Info(fmt.Sprintf("notification sent: event_id=%s, user_id=%s", notification.EventID, notification.UserID))
}
