package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/app"
	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/logger"
	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage/sql"
	"github.com/google/uuid"
	_ "github.com/lib/pq" // PostgreSQL driver
	"gopkg.in/yaml.v3"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/config.yaml", "Path to configuration file")
}

type Config struct {
	Logger  LoggerConf  `yaml:"logger"`
	Server  ServerConf  `yaml:"server"`
	Storage StorageConf `yaml:"storage"`
	DB      DBConf      `yaml:"db"`
}

type LoggerConf struct {
	Level string `yaml:"level"`
}

type ServerConf struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type StorageConf struct {
	Type string `yaml:"type"`
}

type DBConf struct {
	DSN string `yaml:"dsn"`
}

func NewConfigFromFile(path string) (Config, error) {
	if path == "" {
		return Config{}, fmt.Errorf("empty config path")
	}

	f, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(f, &cfg); err != nil {
		return Config{}, err
	}

	// defaults
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Storage.Type == "" {
		cfg.Storage.Type = "memory"
	}

	return cfg, nil
}

func main() {
	flag.Parse()

	cfg, err := NewConfigFromFile(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logg := logger.New(cfg.Logger.Level)

	var store app.Storage
	ctx := context.Background()

	switch cfg.Storage.Type {
	case "sql":
		sql := sqlstorage.New(cfg.DB.DSN)
		if err := sql.Connect(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to db: %v\n", err)
			os.Exit(1)
		}
		defer sql.Close(ctx)
		store = sql
		logg.Info("Using SQL storage")
	default:
		store = memorystorage.New()
		logg.Info("Using memory storage")
	}

	calendarApp := app.New(logg, store)

	fmt.Println("=== Тестирование хранилища календаря ===\n")

	// Создаем тестовые события
	now := time.Now()
	events := []storage.Event{
		{
			ID:           uuid.New().String(),
			Title:        "Встреча с командой",
			At:           now.Add(2 * time.Hour),
			Duration:     1 * time.Hour,
			Description:  "Еженедельная встреча с командой разработки",
			UserID:       "user1",
			NotifyBefore: 15 * time.Minute,
		},
		{
			ID:           uuid.New().String(),
			Title:        "Презентация проекта",
			At:           now.Add(24 * time.Hour),
			Duration:     2 * time.Hour,
			Description:  "Презентация нового проекта клиенту",
			UserID:       "user1",
			NotifyBefore: 30 * time.Minute,
		},
		{
			ID:           uuid.New().String(),
			Title:        "Обед",
			At:           now.Add(6 * time.Hour),
			Duration:     1 * time.Hour,
			Description:  "Обед с коллегами",
			UserID:       "user2",
			NotifyBefore: 0,
		},
		{
			ID:           uuid.New().String(),
			Title:        "Встреча через неделю",
			At:           now.Add(7 * 24 * time.Hour),
			Duration:     30 * time.Minute,
			Description:  "Встреча через неделю",
			UserID:       "user1",
			NotifyBefore: 1 * time.Hour,
		},
		{
			ID:           uuid.New().String(),
			Title:        "Событие в следующем месяце",
			At:           now.AddDate(0, 1, 0),
			Duration:     1 * time.Hour,
			Description:  "Событие в следующем месяце",
			UserID:       "user1",
			NotifyBefore: 24 * time.Hour,
		},
	}

	// Тест 1: Создание событий
	fmt.Println("1. Создание событий:")
	for i, e := range events {
		if err := calendarApp.CreateEvent(ctx, e); err != nil {
			fmt.Printf("   ❌ Ошибка при создании события %d: %v\n", i+1, err)
		} else {
			fmt.Printf("   ✅ Событие %d создано: %s (ID: %s)\n", i+1, e.Title, e.ID)
		}
	}
	fmt.Println()

	// Тест 2: Получение всех событий
	fmt.Println("2. Получение всех событий:")
	allEvents, err := calendarApp.ListEvents(ctx)
	if err != nil {
		fmt.Printf("   ❌ Ошибка при получении списка событий: %v\n", err)
	} else {
		fmt.Printf("   ✅ Найдено событий: %d\n", len(allEvents))
		for _, e := range allEvents {
			fmt.Printf("      - %s (ID: %s, Время: %s)\n", e.Title, e.ID, e.At.Format("2006-01-02 15:04:05"))
		}
	}
	fmt.Println()

	// Тест 3: Получение события по ID
	fmt.Println("3. Получение события по ID:")
	if len(events) > 0 {
		event, err := calendarApp.GetEvent(ctx, events[0].ID)
		if err != nil {
			fmt.Printf("   ❌ Ошибка при получении события: %v\n", err)
		} else {
			fmt.Printf("   ✅ Событие найдено: %s\n", event.Title)
			fmt.Printf("      Описание: %s\n", event.Description)
			fmt.Printf("      Пользователь: %s\n", event.UserID)
			fmt.Printf("      Длительность: %v\n", event.Duration)
		}
	}
	fmt.Println()

	// Тест 4: Обновление события
	fmt.Println("4. Обновление события:")
	if len(events) > 0 {
		updatedEvent := events[0]
		updatedEvent.Title = "Обновленная встреча с командой"
		updatedEvent.Description = "Обновленное описание встречи"
		if err := calendarApp.UpdateEvent(ctx, updatedEvent); err != nil {
			fmt.Printf("   ❌ Ошибка при обновлении события: %v\n", err)
		} else {
			fmt.Printf("   ✅ Событие обновлено\n")
			// Проверяем, что обновление применилось
			event, _ := calendarApp.GetEvent(ctx, updatedEvent.ID)
			fmt.Printf("      Новое название: %s\n", event.Title)
		}
	}
	fmt.Println()

	// Тест 5: Список событий за день
	fmt.Println("5. Список событий за день:")
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEvents, err := calendarApp.ListEventsDay(ctx, today)
	if err != nil {
		fmt.Printf("   ❌ Ошибка при получении событий за день: %v\n", err)
	} else {
		fmt.Printf("   ✅ Событий за сегодня: %d\n", len(dayEvents))
		for _, e := range dayEvents {
			fmt.Printf("      - %s в %s\n", e.Title, e.At.Format("15:04:05"))
		}
	}
	fmt.Println()

	// Тест 6: Список событий за неделю
	fmt.Println("6. Список событий за неделю:")
	weekEvents, err := calendarApp.ListEventsWeek(ctx, today)
	if err != nil {
		fmt.Printf("   ❌ Ошибка при получении событий за неделю: %v\n", err)
	} else {
		fmt.Printf("   ✅ Событий за неделю: %d\n", len(weekEvents))
		for _, e := range weekEvents {
			fmt.Printf("      - %s в %s\n", e.Title, e.At.Format("2006-01-02 15:04:05"))
		}
	}
	fmt.Println()

	// Тест 7: Список событий за месяц
	fmt.Println("7. Список событий за месяц:")
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEvents, err := calendarApp.ListEventsMonth(ctx, monthStart)
	if err != nil {
		fmt.Printf("   ❌ Ошибка при получении событий за месяц: %v\n", err)
	} else {
		fmt.Printf("   ✅ Событий за месяц: %d\n", len(monthEvents))
		for _, e := range monthEvents {
			fmt.Printf("      - %s в %s\n", e.Title, e.At.Format("2006-01-02 15:04:05"))
		}
	}
	fmt.Println()

	// Тест 8: Удаление события
	fmt.Println("8. Удаление события:")
	if len(events) > 1 {
		eventToDelete := events[1]
		if err := calendarApp.DeleteEvent(ctx, eventToDelete.ID); err != nil {
			fmt.Printf("   ❌ Ошибка при удалении события: %v\n", err)
		} else {
			fmt.Printf("   ✅ Событие удалено: %s\n", eventToDelete.Title)
			// Проверяем, что событие действительно удалено
			_, err := calendarApp.GetEvent(ctx, eventToDelete.ID)
			if err != nil {
				fmt.Printf("   ✅ Подтверждено: событие больше не существует\n")
			} else {
				fmt.Printf("   ⚠️  Предупреждение: событие все еще существует\n")
			}
		}
	}
	fmt.Println()

	// Финальная проверка: сколько событий осталось
	fmt.Println("9. Финальная проверка:")
	finalEvents, err := calendarApp.ListEvents(ctx)
	if err != nil {
		fmt.Printf("   ❌ Ошибка: %v\n", err)
	} else {
		fmt.Printf("   ✅ Всего событий в хранилище: %d\n", len(finalEvents))
	}

	fmt.Println("\n=== Тестирование завершено ===")
}
