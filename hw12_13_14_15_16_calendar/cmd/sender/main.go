package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/logger"
	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/rabbitmq"
	_ "github.com/lib/pq"
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

	// Подключаемся к БД (если указан DSN)
	var db *sql.DB
	if cfg.DB.DSN != "" {
		var err error
		db, err = sql.Open("postgres", cfg.DB.DSN)
		if err != nil {
			logg.Error("failed to open database: " + err.Error())
			os.Exit(1)
		}
		defer db.Close()

		// Проверяем подключение
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			logg.Error("failed to ping database: " + err.Error())
			os.Exit(1)
		}
		logg.Info("connected to database")
	}

	// Подключаемся к RabbitMQ с retry
	rmqClient := rabbitmq.NewClient(cfg.RabbitMQ.URL)
	ctxConnect, cancelConnect := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelConnect()
	
	maxRetries := 10
	retryInterval := 3 * time.Second
	for i := 0; i < maxRetries; i++ {
		if err := rmqClient.Connect(ctxConnect); err == nil {
			break
		}
		if i < maxRetries-1 {
			logg.Debug(fmt.Sprintf("failed to connect to RabbitMQ (attempt %d/%d), retrying in %v...", i+1, maxRetries, retryInterval))
			time.Sleep(retryInterval)
		} else {
			logg.Error("failed to connect to RabbitMQ after " + fmt.Sprintf("%d attempts", maxRetries))
			os.Exit(1)
		}
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
			if err := sendNotification(ctx, logg, db, notification); err != nil {
				logg.Error("failed to send notification: " + err.Error())
				// Отклоняем сообщение с повторной постановкой в очередь
				if err := msg.Nack(); err != nil {
					logg.Error("failed to nack message: " + err.Error())
				}
				continue
			}

			// Подтверждаем обработку сообщения
			if err := msg.Ack(); err != nil {
				logg.Error("failed to ack message: " + err.Error())
			}
		}
	}
}

func sendNotification(ctx context.Context, logg *logger.Logger, db *sql.DB, notification *rabbitmq.Notification) error {
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

	// Сохраняем статус уведомления в БД, если подключение есть
	if db != nil {
		_, err := db.ExecContext(ctx, `
			INSERT INTO notifications (event_id, user_id, title, event_date, status, sent_at)
			VALUES ($1, $2, $3, $4, 'sent', NOW())
		`, notification.EventID, notification.UserID, notification.Title, notification.EventDate)
		if err != nil {
			return fmt.Errorf("failed to save notification status: %w", err)
		}
	}

	return nil
}
