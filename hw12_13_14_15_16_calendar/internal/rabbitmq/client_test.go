package rabbitmq

import (
	"context"
	"testing"
	"time"
)

func TestNotification_Serialize(t *testing.T) {
	n := &Notification{
		EventID:   "test-id",
		Title:     "Test Event",
		EventDate: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		UserID:    "user-123",
	}

	data, err := n.Serialize()
	if err != nil {
		t.Fatalf("failed to serialize notification: %v", err)
	}

	if len(data) == 0 {
		t.Error("serialized data is empty")
	}

	// Проверяем, что можно десериализовать обратно
	deserialized, err := DeserializeNotification(data)
	if err != nil {
		t.Fatalf("failed to deserialize notification: %v", err)
	}

	if deserialized.EventID != n.EventID {
		t.Errorf("expected EventID %s, got %s", n.EventID, deserialized.EventID)
	}
	if deserialized.Title != n.Title {
		t.Errorf("expected Title %s, got %s", n.Title, deserialized.Title)
	}
	if deserialized.UserID != n.UserID {
		t.Errorf("expected UserID %s, got %s", n.UserID, deserialized.UserID)
	}
	if !deserialized.EventDate.Equal(n.EventDate) {
		t.Errorf("expected EventDate %v, got %v", n.EventDate, deserialized.EventDate)
	}
}

func TestNotification_Validate(t *testing.T) {
	tests := []struct {
		name         string
		notification *Notification
		wantErr      bool
	}{
		{
			name: "valid notification",
			notification: &Notification{
				EventID:   "test-id",
				Title:     "Test Event",
				EventDate: time.Now(),
				UserID:    "user-123",
			},
			wantErr: false,
		},
		{
			name: "missing event_id",
			notification: &Notification{
				Title:     "Test Event",
				EventDate: time.Now(),
				UserID:    "user-123",
			},
			wantErr: true,
		},
		{
			name: "missing title",
			notification: &Notification{
				EventID:   "test-id",
				EventDate: time.Now(),
				UserID:    "user-123",
			},
			wantErr: true,
		},
		{
			name: "missing user_id",
			notification: &Notification{
				EventID:   "test-id",
				Title:     "Test Event",
				EventDate: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.notification.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeserializeNotification_InvalidJSON(t *testing.T) {
	invalidData := []byte("invalid json")
	_, err := DeserializeNotification(invalidData)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// MockClient для тестирования без реального RabbitMQ
type MockClient struct {
	publishedMessages []PublishedMessage
	consumedMessages  []Message
	connectError      error
	publishError      error
	consumeError      error
}

type PublishedMessage struct {
	QueueName string
	Message   []byte
}

func NewMockClient() *MockClient {
	return &MockClient{
		publishedMessages: make([]PublishedMessage, 0),
		consumedMessages:  make([]Message, 0),
	}
}

func (m *MockClient) Connect(ctx context.Context) error {
	return m.connectError
}

func (m *MockClient) Close() error {
	return nil
}

func (m *MockClient) Publish(ctx context.Context, queueName string, message []byte) error {
	if m.publishError != nil {
		return m.publishError
	}
	m.publishedMessages = append(m.publishedMessages, PublishedMessage{
		QueueName: queueName,
		Message:   message,
	})
	return nil
}

func (m *MockClient) Consume(ctx context.Context, queueName string) (<-chan Message, error) {
	if m.consumeError != nil {
		return nil, m.consumeError
	}
	ch := make(chan Message, 1)
	// Добавляем сообщения в канал
	for _, msg := range m.consumedMessages {
		ch <- msg
	}
	close(ch)
	return ch, nil
}

func (m *MockClient) AddConsumedMessage(msg Message) {
	m.consumedMessages = append(m.consumedMessages, msg)
}

func TestMockClient_Publish(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	notification := &Notification{
		EventID:   "test-id",
		Title:     "Test Event",
		EventDate: time.Now(),
		UserID:    "user-123",
	}

	message, err := notification.Serialize()
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	err = mock.Publish(ctx, "test-queue", message)
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	if len(mock.publishedMessages) != 1 {
		t.Fatalf("expected 1 published message, got %d", len(mock.publishedMessages))
	}

	published := mock.publishedMessages[0]
	if published.QueueName != "test-queue" {
		t.Errorf("expected queue name 'test-queue', got '%s'", published.QueueName)
	}

	// Проверяем, что можно десериализовать
	deserialized, err := DeserializeNotification(published.Message)
	if err != nil {
		t.Fatalf("failed to deserialize published message: %v", err)
	}

	if deserialized.EventID != notification.EventID {
		t.Errorf("expected EventID %s, got %s", notification.EventID, deserialized.EventID)
	}
}

func TestMockClient_Consume(t *testing.T) {
	mock := NewMockClient()
	ctx := context.Background()

	notification := &Notification{
		EventID:   "test-id",
		Title:     "Test Event",
		EventDate: time.Now(),
		UserID:    "user-123",
	}

	message, err := notification.Serialize()
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Добавляем сообщение для потребления
	mock.AddConsumedMessage(Message{
		Body: message,
		Ack: func() error {
			return nil
		},
		Nack: func() error {
			return nil
		},
		Reject: func() error {
			return nil
		},
	})

	ch, err := mock.Consume(ctx, "test-queue")
	if err != nil {
		t.Fatalf("failed to consume: %v", err)
	}

	msg := <-ch
	if msg.Body == nil {
		t.Error("received message body is nil")
	}

	deserialized, err := DeserializeNotification(msg.Body)
	if err != nil {
		t.Fatalf("failed to deserialize consumed message: %v", err)
	}

	if deserialized.EventID != notification.EventID {
		t.Errorf("expected EventID %s, got %s", notification.EventID, deserialized.EventID)
	}
}
