package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// Client представляет интерфейс для работы с RabbitMQ
// Процессы не должны зависеть от конкретной реализации RMQ-клиента
type Client interface {
	// Connect подключается к RabbitMQ и создает все необходимые структуры
	Connect(ctx context.Context) error
	// Close закрывает соединение
	Close() error
	// Publish отправляет сообщение в очередь
	Publish(ctx context.Context, queueName string, message []byte) error
	// Consume начинает чтение сообщений из очереди
	Consume(ctx context.Context, queueName string) (<-chan Message, error)
}

// Message представляет сообщение из очереди
type Message struct {
	Body   []byte
	Ack    func() error
	Nack   func() error
	Reject func() error
}

// RabbitMQClient - реализация Client на основе amqp091-go
type RabbitMQClient struct {
	url     string
	conn    *amqp091.Connection
	channel *amqp091.Channel
	queues  map[string]bool // отслеживание созданных очередей
}

// NewClient создает новый клиент RabbitMQ
func NewClient(url string) *RabbitMQClient {
	return &RabbitMQClient{
		url:    url,
		queues: make(map[string]bool),
	}
}

// Connect подключается к RabbitMQ
func (c *RabbitMQClient) Connect(ctx context.Context) error {
	var err error
	c.conn, err = amqp091.Dial(c.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	return nil
}

// Close закрывает соединение
func (c *RabbitMQClient) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ensureQueue создает очередь, если она еще не создана
func (c *RabbitMQClient) ensureQueue(queueName string) error {
	if c.queues[queueName] {
		return nil
	}

	_, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	c.queues[queueName] = true
	return nil
}

// Publish отправляет сообщение в очередь
func (c *RabbitMQClient) Publish(ctx context.Context, queueName string, message []byte) error {
	if err := c.ensureQueue(queueName); err != nil {
		return err
	}

	err := c.channel.PublishWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         message,
			DeliveryMode: amqp091.Persistent, // сообщения сохраняются на диск
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Consume начинает чтение сообщений из очереди
func (c *RabbitMQClient) Consume(ctx context.Context, queueName string) (<-chan Message, error) {
	if err := c.ensureQueue(queueName); err != nil {
		return nil, err
	}

	// Настраиваем QoS для fair dispatch
	err := c.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := c.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack (false - ручное подтверждение)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}

	messageChan := make(chan Message)
	go func() {
		defer close(messageChan)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				messageChan <- Message{
					Body: msg.Body,
					Ack: func() error {
						return msg.Ack(false)
					},
					Nack: func() error {
						return msg.Nack(false, true) // requeue
					},
					Reject: func() error {
						return msg.Nack(false, false) // не requeue
					},
				}
			}
		}
	}()

	return messageChan, nil
}

// Notification представляет уведомление о событии
type Notification struct {
	EventID   string    `json:"event_id"`
	Title     string    `json:"title"`
	EventDate time.Time `json:"event_date"`
	UserID    string    `json:"user_id"`
}

// Serialize сериализует уведомление в JSON
func (n *Notification) Serialize() ([]byte, error) {
	return json.Marshal(n)
}

// DeserializeNotification десериализует уведомление из JSON
func DeserializeNotification(data []byte) (*Notification, error) {
	var n Notification
	if err := json.Unmarshal(data, &n); err != nil {
		return nil, fmt.Errorf("failed to deserialize notification: %w", err)
	}
	return &n, nil
}

// Validate проверяет корректность уведомления
func (n *Notification) Validate() error {
	if n.EventID == "" {
		return errors.New("event_id is required")
	}
	if n.Title == "" {
		return errors.New("title is required")
	}
	if n.UserID == "" {
		return errors.New("user_id is required")
	}
	return nil
}
