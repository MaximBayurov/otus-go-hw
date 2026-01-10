package contracts

import "context"

// MessageHandler - обработчик сообщений из очереди.
type MessageHandler func(ctx context.Context, body []byte) error

// Client - интерфейс для работы с Broker.
type Client interface {
	// Connect подключение к брокеру
	Connect(ctx context.Context) error

	// Close завершение подключения
	Close() error

	// Publish публикация сообщения
	Publish(ctx context.Context, exchange, routingKey string, message []byte) error

	// PublishWithRetry публикация сообщения с повторной отправкой
	PublishWithRetry(ctx context.Context, exchange, routingKey string, message []byte, maxRetries int) error

	// Consume подписка на сообщения
	Consume(ctx context.Context, queue string, handler MessageHandler) error
}
