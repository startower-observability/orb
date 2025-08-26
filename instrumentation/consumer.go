package instrumentation

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/startower-observability/orb/internal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type ConsumerConfig struct {
	Tracer            trace.Tracer
	Propagator        *Propagator
	SpanNameFormatter func(queueName string, delivery *amqp091.Delivery) string
	AttributeEnricher func(ctx context.Context, queueName string, delivery *amqp091.Delivery) []trace.SpanStartOption
}

type Consumer struct {
	config ConsumerConfig
}

func NewConsumer(config ConsumerConfig) *Consumer {
	if config.Tracer == nil {
		config.Tracer = otel.Tracer(internal.TracerName)
	}
	if config.Propagator == nil {
		config.Propagator = DefaultPropagator
	}
	if config.SpanNameFormatter == nil {
		config.SpanNameFormatter = defaultConsumeSpanName
	}

	return &Consumer{
		config: config,
	}
}

func NewDefaultConsumer() *Consumer {
	return NewConsumer(ConsumerConfig{})
}

type MessageHandler func(ctx context.Context, delivery amqp091.Delivery) error

func (c *Consumer) ConsumeWithHandler(
	ctx context.Context,
	channel *amqp091.Channel,
	queueName, consumerTag string,
	autoAck, exclusive, noLocal, noWait bool,
	args amqp091.Table,
	handler MessageHandler,
) error {
	deliveries, err := channel.ConsumeWithContext(
		ctx, queueName, consumerTag, autoAck, exclusive, noLocal, noWait, args,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	go func() {
		for delivery := range deliveries {
			c.processDelivery(ctx, queueName, delivery, handler, autoAck)
		}
	}()

	return nil
}

func (c *Consumer) ProcessDelivery(
	ctx context.Context,
	queueName string,
	delivery amqp091.Delivery,
	handler MessageHandler,
) error {
	c.processDelivery(ctx, queueName, delivery, handler, false)
	return nil
}

func (c *Consumer) processDelivery(
	parentCtx context.Context,
	queueName string,
	delivery amqp091.Delivery,
	handler MessageHandler,
	autoAck bool,
) {
	ctx := c.config.Propagator.ExtractFromDelivery(parentCtx, &delivery)

	spanName := c.config.SpanNameFormatter(queueName, &delivery)

	spanOpts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindConsumer),
	}

	attrs := internal.GetConsumeAttributes(queueName, &delivery)
	for _, attr := range attrs {
		spanOpts = append(spanOpts, trace.WithAttributes(attr))
	}

	if c.config.AttributeEnricher != nil {
		customOpts := c.config.AttributeEnricher(ctx, queueName, &delivery)
		spanOpts = append(spanOpts, customOpts...)
	}

	ctx, span := c.config.Tracer.Start(ctx, spanName, spanOpts...)
	defer span.End()

	var err error
	if handler != nil {
		err = handler(ctx, delivery)
	}

	if !autoAck {
		if err != nil {
			if nackErr := delivery.Nack(false, true); nackErr != nil {
				span.RecordError(fmt.Errorf("failed to nack message: %w", nackErr))
			}
		} else {
			if ackErr := delivery.Ack(false); ackErr != nil {
				span.RecordError(fmt.Errorf("failed to ack message: %w", ackErr))
				err = ackErr
			}
		}
	}

	internal.SafeSetSpanStatus(span, err)
}

func (c *Consumer) WrapDelivery(
	ctx context.Context,
	queueName string,
	delivery *amqp091.Delivery,
) (context.Context, trace.Span) {
	ctx = c.config.Propagator.ExtractFromDelivery(ctx, delivery)

	spanName := c.config.SpanNameFormatter(queueName, delivery)

	spanOpts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindConsumer),
	}

	attrs := internal.GetConsumeAttributes(queueName, delivery)
	for _, attr := range attrs {
		spanOpts = append(spanOpts, trace.WithAttributes(attr))
	}

	if c.config.AttributeEnricher != nil {
		customOpts := c.config.AttributeEnricher(ctx, queueName, delivery)
		spanOpts = append(spanOpts, customOpts...)
	}

	return c.config.Tracer.Start(ctx, spanName, spanOpts...)
}

func defaultConsumeSpanName(queueName string, delivery *amqp091.Delivery) string {
	if queueName != "" {
		return fmt.Sprintf("%s receive", queueName)
	}
	if delivery.RoutingKey != "" {
		return fmt.Sprintf("%s receive", delivery.RoutingKey)
	}
	return "rabbitmq receive"
}

var defaultConsumer = NewDefaultConsumer()

func ConsumeWithHandler(
	ctx context.Context,
	channel *amqp091.Channel,
	queueName, consumerTag string,
	autoAck, exclusive, noLocal, noWait bool,
	args amqp091.Table,
	handler MessageHandler,
) error {
	return defaultConsumer.ConsumeWithHandler(
		ctx, channel, queueName, consumerTag, autoAck, exclusive, noLocal, noWait, args, handler,
	)
}

func ProcessDelivery(
	ctx context.Context,
	queueName string,
	delivery amqp091.Delivery,
	handler MessageHandler,
) error {
	return defaultConsumer.ProcessDelivery(ctx, queueName, delivery, handler)
}

func WrapDelivery(
	ctx context.Context,
	queueName string,
	delivery *amqp091.Delivery,
) (context.Context, trace.Span) {
	return defaultConsumer.WrapDelivery(ctx, queueName, delivery)
}
