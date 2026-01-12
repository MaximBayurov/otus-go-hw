package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/broker/contracts"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/streadway/amqp"
)

const maxRetryCount = 3

func NewRabbitMQClient(config configuration.BrokerConf, logger *logger.Logger) *Client {
	return &Client{
		config: config,
		logger: logger,
	}
}

// Client - реализация клиента Broker.
type Client struct {
	config     configuration.BrokerConf
	connection *amqp.Connection
	channel    *amqp.Channel
	logger     *logger.Logger
}

// Connect устанавливает соединение с Broker и создает необходимые структуры.
func (c *Client) Connect(_ context.Context) error {
	c.logger.Info("start connect to broker: " + c.config.URL())

	var err error
	c.connection, err = amqp.Dial(c.config.URL())
	if err != nil {
		return fmt.Errorf("connecting to broker: %w", err)
	}

	c.channel, err = c.connection.Channel()
	if err != nil {
		return fmt.Errorf("chanel create: %w", err)
	}

	err = c.channel.ExchangeDeclare(
		c.config.ExchangeName,
		c.config.ExchangeType,
		c.config.Durable,
		c.config.AutoDelete,
		false,
		c.config.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("exchange declare: %w", err)
	}

	queue, err := c.channel.QueueDeclare(
		c.config.QueueName,
		c.config.Durable,
		c.config.AutoDelete,
		c.config.Exclusive,
		c.config.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("queue declare: %w", err)
	}

	err = c.channel.QueueBind(
		queue.Name,
		c.config.RoutingKey,
		c.config.ExchangeName,
		c.config.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("queue bind: %w", err)
	}

	c.logger.Info(
		"successfully connect to broker." +
			fmt.Sprintf("\nexchange: %s", c.config.ExchangeName) +
			fmt.Sprintf("\nqueue: %s", queue.Name) +
			fmt.Sprintf("\nmessages: %d", queue.Messages),
	)
	return nil
}

// Close закрывает соединение.
func (c *Client) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			c.logger.Error(fmt.Sprintf("channel close error: %s", err.Error()))
		}
	}

	if c.connection != nil {
		if err := c.connection.Close(); err != nil {
			c.logger.Error(fmt.Sprintf("connection close error: %s", err.Error()))
			return err
		}
	}

	c.logger.Info("Broker connection closed")
	return nil
}

// Publish публикует сообщение в очередь.
func (c *Client) Publish(ctx context.Context, exchange, routingKey string, message []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if c.channel == nil {
		return fmt.Errorf("channel doesnt initialized")
	}

	err := c.channel.Publish(
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         message,
			DeliveryMode: amqp.Persistent, // Сохраняем на диск
			Timestamp:    time.Now(),
			Headers: amqp.Table{
				"x-retry-count": 0,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("message publishing: %w", err)
	}

	c.logger.Debug(
		"message is published." +
			fmt.Sprintf("\nexchange: %s", c.config.ExchangeName) +
			fmt.Sprintf("\nrouting_key: %s", routingKey) +
			fmt.Sprintf("\nsize: %d", len(message)),
	)

	return nil
}

// PublishWithRetry публикует сообщение с повторными попытками.
func (c *Client) PublishWithRetry(
	ctx context.Context,
	exchange,
	routingKey string,
	message []byte,
	maxRetries int,
) error {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := c.Publish(ctx, exchange, routingKey, message)
		if err == nil {
			return nil
		}

		lastErr = err

		c.logger.Debug(
			"message publishing error, retrying: " +
				fmt.Sprintf("\nattempt: %d", i+1) +
				fmt.Sprintf("\nmax_retries: %d", maxRetries) +
				fmt.Sprintf("\nerror: %s", err.Error()),
		)

		if i < maxRetries {
			time.Sleep(time.Duration(i+1) * time.Second) // Exponential backoff
		}
	}

	return fmt.Errorf("failed to publish after %d attempts: %w", maxRetries, lastErr)
}

// Consume начинает потребление сообщений из очереди.
func (c *Client) Consume(ctx context.Context, queue string, handler contracts.MessageHandler) error {
	messagesChannel, err := c.channel.Consume(
		queue,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("message consumption: %w", err)
	}

	c.logger.Info(
		"start message consumption." +
			fmt.Sprintf("\nqueue: %s", queue),
	)

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Info(
					"stop message consumption." +
						fmt.Sprintf("\nqueue: %s", queue),
				)
				return
			case msg, ok := <-messagesChannel:
				if !ok {
					c.logger.Info(
						"channel closed. " +
							fmt.Sprintf("\nqueue: %s", queue),
					)

					return
				}

				c.processMessage(ctx, msg, handler)
			}
		}
	}()

	return nil
}

// processMessage обрабатывает одно сообщение.
func (c *Client) processMessage(ctx context.Context, msg amqp.Delivery, handler contracts.MessageHandler) {
	c.logger.Debug(
		"message received: " +
			fmt.Sprintf("\nmessage_id: %s", msg.MessageId) +
			fmt.Sprintf("\ncorrelation_id: %s", msg.CorrelationId) +
			fmt.Sprintf("\ntimestamp: %s", msg.Timestamp),
	)

	// Создаем контекст с таймаутом для обработки
	processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Обрабатываем сообщение
	err := handler(processCtx, msg.Body)
	if err != nil {
		c.logger.Debug(
			"message handling error: " +
				fmt.Sprintf("\nerror: %s", err.Error()) +
				fmt.Sprintf("\nmessage_id: %s", msg.MessageId),
		)

		// Проверяем количество попыток
		retryCount, _ := msg.Headers["x-retry-count"].(int32)
		if retryCount < maxRetryCount {
			// Повторная публикация с увеличенным счетчиком
			msg.Headers["x-retry-count"] = retryCount + 1
			msg.Headers["x-retry-reason"] = err.Error()

			// Задержка перед повторной попыткой
			delay := time.Duration(retryCount+1) * 5 * time.Second
			time.Sleep(delay)

			c.channel.Publish(
				"",
				msg.RoutingKey,
				false,
				false,
				amqp.Publishing{
					ContentType:  msg.ContentType,
					Body:         msg.Body,
					Headers:      msg.Headers,
					DeliveryMode: amqp.Persistent,
					Timestamp:    time.Now(),
				},
			)
		} else {
			c.logger.Error(
				fmt.Sprintf("message moved to DLQ after %d attempts.", maxRetryCount) +
					fmt.Sprintf("\nmessage_id: %s", msg.MessageId),
			)

			// Отправляем в Dead Letter Queue
			c.channel.Publish(
				"dlx",
				"dlq."+msg.RoutingKey,
				false,
				false,
				amqp.Publishing{
					ContentType:  msg.ContentType,
					Body:         msg.Body,
					Headers:      msg.Headers,
					DeliveryMode: amqp.Persistent,
					Timestamp:    time.Now(),
				},
			)
		}

		// Nack сообщения
		msg.Nack(false, false)
		return
	}

	// Ack сообщения при успешной обработке
	msg.Ack(false)

	c.logger.Debug(
		"message successfully handled." +
			fmt.Sprintf("\nmessage_id: %s", msg.MessageId),
	)
}
