package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/app"
	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/logger"
	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/rabbitmq"
	sqlstorage "github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage/sql"
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

	// Подключаемся к БД
	storage := sqlstorage.New(cfg.DB.DSN)
	if err := storage.Connect(context.Background()); err != nil {
		logg.Error("failed to connect to db: " + err.Error())
		os.Exit(1)
	}
	defer storage.Close(context.Background())

	calendar := app.New(logg, storage)

	// Подключаемся к RabbitMQ
	rmqClient := rabbitmq.NewClient(cfg.RabbitMQ.URL)
	if err := rmqClient.Connect(context.Background()); err != nil {
		logg.Error("failed to connect to RabbitMQ: " + err.Error())
		os.Exit(1)
	}
	defer rmqClient.Close()

	logg.Info("scheduler is running...")

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	// Запускаем периодическое сканирование
	ticker := time.NewTicker(cfg.Scheduler.Interval)
	defer ticker.Stop()

	// Выполняем первую итерацию сразу
	if err := processEvents(ctx, logg, calendar, rmqClient, cfg.RabbitMQ.QueueName); err != nil {
		logg.Error("failed to process events: " + err.Error())
	}

	for {
		select {
		case <-ctx.Done():
			logg.Info("scheduler is stopping...")
			return
		case <-ticker.C:
			if err := processEvents(ctx, logg, calendar, rmqClient, cfg.RabbitMQ.QueueName); err != nil {
				logg.Error("failed to process events: " + err.Error())
			}
		}
	}
}

func processEvents(ctx context.Context, logg app.Logger, calendar *app.App, rmqClient rabbitmq.Client, queueName string) error {
	now := time.Now()

	// Получаем события, требующие уведомления
	events, err := calendar.EventsToNotify(ctx, now)
	if err != nil {
		return fmt.Errorf("failed to get events to notify: %w", err)
	}

	logg.Debug(fmt.Sprintf("found %d events to notify", len(events)))

	// Отправляем уведомления в очередь
	for _, event := range events {
		notification := &rabbitmq.Notification{
			EventID:   event.ID,
			Title:     event.Title,
			EventDate: event.At,
			UserID:    event.UserID,
		}

		if err := notification.Validate(); err != nil {
			logg.Error(fmt.Sprintf("invalid notification for event %s: %v", event.ID, err))
			continue
		}

		message, err := notification.Serialize()
		if err != nil {
			logg.Error(fmt.Sprintf("failed to serialize notification for event %s: %v", event.ID, err))
			continue
		}

		if err := rmqClient.Publish(ctx, queueName, message); err != nil {
			logg.Error(fmt.Sprintf("failed to publish notification for event %s: %v", event.ID, err))
			continue
		}

		logg.Info(fmt.Sprintf("notification sent for event %s (user: %s, date: %s)", event.ID, event.UserID, event.At.Format(time.RFC3339)))
	}

	// Удаляем старые события (более 1 года назад)
	oneYearAgo := now.AddDate(-1, 0, 0)
	deletedCount, err := calendar.DeleteOldEvents(ctx, oneYearAgo)
	if err != nil {
		return fmt.Errorf("failed to delete old events: %w", err)
	}

	if deletedCount > 0 {
		logg.Info(fmt.Sprintf("deleted %d old events (before %s)", deletedCount, oneYearAgo.Format(time.RFC3339)))
	}

	return nil
}
