package instrumentation

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
)

type Channel struct {
	*amqp091.Channel
	publisher *Publisher
	consumer  *Consumer
}

type ChannelConfig struct {
	PublisherConfig PublisherConfig
	ConsumerConfig  ConsumerConfig
}

func NewChannel(channel *amqp091.Channel, config ChannelConfig) *Channel {
	return &Channel{
		Channel:   channel,
		publisher: NewPublisher(config.PublisherConfig),
		consumer:  NewConsumer(config.ConsumerConfig),
	}
}

func NewDefaultChannel(channel *amqp091.Channel) *Channel {
	return NewChannel(channel, ChannelConfig{})
}

func (c *Channel) PublishWithTracing(
	ctx context.Context,
	exchange, routingKey string,
	mandatory, immediate bool,
	msg amqp091.Publishing,
) error {
	return c.publisher.Publish(ctx, c.Channel, exchange, routingKey, mandatory, immediate, msg)
}

func (c *Channel) PublishWithConfirmAndTracing(
	ctx context.Context,
	exchange, routingKey string,
	mandatory, immediate bool,
	msg amqp091.Publishing,
) (*amqp091.DeferredConfirmation, error) {
	return c.publisher.PublishWithConfirm(ctx, c.Channel, exchange, routingKey, mandatory, immediate, msg)
}

func (c *Channel) ConsumeWithTracing(
	ctx context.Context,
	queueName, consumerTag string,
	autoAck, exclusive, noLocal, noWait bool,
	args amqp091.Table,
	handler MessageHandler,
) error {
	return c.consumer.ConsumeWithHandler(
		ctx, c.Channel, queueName, consumerTag, autoAck, exclusive, noLocal, noWait, args, handler,
	)
}

func (c *Channel) ProcessDeliveryWithTracing(
	ctx context.Context,
	queueName string,
	delivery amqp091.Delivery,
	handler MessageHandler,
) error {
	return c.consumer.ProcessDelivery(ctx, queueName, delivery, handler)
}

func (c *Channel) WrapDeliveryWithTracing(
	ctx context.Context,
	queueName string,
	delivery *amqp091.Delivery,
) (context.Context, trace.Span) {
	return c.consumer.WrapDelivery(ctx, queueName, delivery)
}

func (c *Channel) GetPublisher() *Publisher {
	return c.publisher
}

func (c *Channel) GetConsumer() *Consumer {
	return c.consumer
}

type Connection struct {
	*amqp091.Connection
	channelConfig ChannelConfig
}

type ConnectionConfig struct {
	ChannelConfig ChannelConfig
}

func NewConnection(conn *amqp091.Connection, config ConnectionConfig) *Connection {
	return &Connection{
		Connection:    conn,
		channelConfig: config.ChannelConfig,
	}
}

func NewDefaultConnection(conn *amqp091.Connection) *Connection {
	return NewConnection(conn, ConnectionConfig{})
}

func (c *Connection) ChannelWithTracing() (*Channel, error) {
	ch, err := c.Connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}
	return NewChannel(ch, c.channelConfig), nil
}

func (c *Connection) ChannelWithTracingAndConfig(config ChannelConfig) (*Channel, error) {
	ch, err := c.Connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}
	return NewChannel(ch, config), nil
}

func Dial(url string) (*Connection, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	return NewDefaultConnection(conn), nil
}

func DialWithConfig(url string, config ConnectionConfig) (*Connection, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	return NewConnection(conn, config), nil
}

func DialConfig(url string, amqpConfig amqp091.Config) (*Connection, error) {
	conn, err := amqp091.DialConfig(url, amqpConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ with config: %w", err)
	}
	return NewDefaultConnection(conn), nil
}

func DialConfigWithConfig(url string, amqpConfig amqp091.Config, config ConnectionConfig) (*Connection, error) {
	conn, err := amqp091.DialConfig(url, amqpConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ with config: %w", err)
	}
	return NewConnection(conn, config), nil
}
