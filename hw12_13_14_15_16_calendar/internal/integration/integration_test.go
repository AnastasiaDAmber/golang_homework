//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL      = "http://calendar:8888"
	testDBDSN    = "postgres://calendar:calendar@db:5432/calendar?sslmode=disable"
	waitTimeout  = 30 * time.Second
	pollInterval = 500 * time.Millisecond
)

func TestMain(m *testing.M) {
	// Ждем, пока сервисы запустятся
	if err := waitForService(baseURL, waitTimeout); err != nil {
		fmt.Printf("Failed to wait for calendar service: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

func waitForService(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(pollInterval)
	}
	return fmt.Errorf("service not available after %v", timeout)
}

func TestIntegration_CreateEvent(t *testing.T) {
	// Тест 1: Успешное создание события
	t.Run("successful creation", func(t *testing.T) {
		event := map[string]interface{}{
			"id":          "test-event-1",
			"title":       "Test Event",
			"at":          time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			"duration":    "1h",
			"description": "Test description",
			"user_id":     "user-1",
		}

		body, _ := json.Marshal(event)
		resp, err := http.Post(baseURL+"/api/events", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Проверяем, что событие создано
		getResp, err := http.Get(baseURL + "/api/events/get?id=test-event-1")
		require.NoError(t, err)
		defer getResp.Body.Close()
		assert.Equal(t, http.StatusOK, getResp.StatusCode)
	})

	// Тест 2: Обработка бизнес ошибок - дублирование события
	t.Run("duplicate event", func(t *testing.T) {
		event := map[string]interface{}{
			"id":      "test-event-duplicate",
			"title":   "Duplicate Event",
			"at":      time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			"user_id": "user-1",
		}

		body, _ := json.Marshal(event)
		resp1, err := http.Post(baseURL+"/api/events", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		resp1.Body.Close()
		assert.Equal(t, http.StatusOK, resp1.StatusCode)

		// Пытаемся создать событие с тем же ID
		body2, _ := json.Marshal(event)
		resp2, err := http.Post(baseURL+"/api/events", "application/json", bytes.NewBuffer(body2))
		require.NoError(t, err)
		defer resp2.Body.Close()

		// Должна быть ошибка
		assert.True(t, resp2.StatusCode >= 400, "expected error status for duplicate event")
	})

	// Тест 3: Валидация - отсутствие обязательных полей
	t.Run("missing required fields", func(t *testing.T) {
		event := map[string]interface{}{
			"id": "test-event-invalid",
			// отсутствует title
		}

		body, _ := json.Marshal(event)
		resp, err := http.Post(baseURL+"/api/events", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Должна быть ошибка валидации
		assert.True(t, resp.StatusCode >= 400, "expected error status for invalid event")
	})
}

func TestIntegration_ListEvents(t *testing.T) {
	now := time.Now().UTC()

	// Создаем тестовые события
	events := []map[string]interface{}{
		{
			"id":      "event-day-1",
			"title":   "Event Today",
			"at":      now.Add(2 * time.Hour).Format(time.RFC3339),
			"user_id": "user-1",
		},
		{
			"id":      "event-day-2",
			"title":   "Event Tomorrow",
			"at":      now.Add(26 * time.Hour).Format(time.RFC3339),
			"user_id": "user-1",
		},
		{
			"id":      "event-week-1",
			"title":   "Event Next Week",
			"at":      now.Add(8 * 24 * time.Hour).Format(time.RFC3339),
			"user_id": "user-1",
		},
		{
			"id":      "event-month-1",
			"title":   "Event Next Month",
			"at":      now.Add(35 * 24 * time.Hour).Format(time.RFC3339),
			"user_id": "user-1",
		},
	}

	// Создаем события
	for _, event := range events {
		body, _ := json.Marshal(event)
		resp, err := http.Post(baseURL+"/api/events", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Тест: Листинг событий на день
	t.Run("list events by day", func(t *testing.T) {
		dayStart := now.Truncate(24 * time.Hour)
		url := fmt.Sprintf("%s/api/events/day?day_start=%s", baseURL, dayStart.Format(time.RFC3339))
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var eventsList []interface{}
		err = json.NewDecoder(resp.Body).Decode(&eventsList)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(eventsList), 1, "expected at least 1 event for today")
	})

	// Тест: Листинг событий на неделю
	t.Run("list events by week", func(t *testing.T) {
		weekStart := now.Truncate(24 * time.Hour)
		url := fmt.Sprintf("%s/api/events/week?week_start=%s", baseURL, weekStart.Format(time.RFC3339))
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var eventsList []interface{}
		err = json.NewDecoder(resp.Body).Decode(&eventsList)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(eventsList), 2, "expected at least 2 events for week")
	})

	// Тест: Листинг событий на месяц
	t.Run("list events by month", func(t *testing.T) {
		monthStart := now.Truncate(24 * time.Hour)
		url := fmt.Sprintf("%s/api/events/month?month_start=%s", baseURL, monthStart.Format(time.RFC3339))
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var eventsList []interface{}
		err = json.NewDecoder(resp.Body).Decode(&eventsList)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(eventsList), 3, "expected at least 3 events for month")
	})
}

func TestIntegration_Notifications(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()

	// Создаем событие с уведомлением
	// Событие через 5 минут, уведомление за 4 минуты
	// Это означает, что время уведомления уже наступило (now >= At - NotifyBefore)
	event := map[string]interface{}{
		"id":            "event-notify-1",
		"title":         "Event with Notification",
		"at":            now.Add(5 * time.Minute).Format(time.RFC3339),
		"user_id":       "user-1",
		"notify_before": "4m",
	}

	body, _ := json.Marshal(event)
	resp, err := http.Post(baseURL+"/api/events", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Ждем, пока scheduler обработает событие и отправит уведомление
	// Scheduler работает с интервалом 1 минута в тестовом окружении
	time.Sleep(10 * time.Second)

	// Проверяем, что уведомление было отправлено (проверяем в БД)
	db, err := sql.Open("postgres", testDBDSN)
	require.NoError(t, err)
	defer db.Close()

	// Ждем, пока уведомление появится в БД (scheduler может обработать с задержкой)
	var notificationCount int
	maxWait := 60 * time.Second
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		err := db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM notifications 
			WHERE event_id = $1 AND status = 'sent'
		`, "event-notify-1").Scan(&notificationCount)
		if err == nil && notificationCount > 0 {
			break
		}
		time.Sleep(2 * time.Second)
	}

	assert.Greater(t, notificationCount, 0, "expected notification to be sent and saved in database")
}
