package instrumentation

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/startower-observability/orb/internal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type PublisherConfig struct {
	Tracer            trace.Tracer
	Propagator        *Propagator
	SpanNameFormatter func(exchange, routingKey string) string
	AttributeEnricher func(ctx context.Context, exchange, routingKey string, msg *amqp091.Publishing) []trace.SpanStartOption
}

type Publisher struct {
	config PublisherConfig
}

func NewPublisher(config PublisherConfig) *Publisher {
	if config.Tracer == nil {
		config.Tracer = otel.Tracer(internal.TracerName)
	}
	if config.Propagator == nil {
		config.Propagator = DefaultPropagator
	}
	if config.SpanNameFormatter == nil {
		config.SpanNameFormatter = defaultPublishSpanName
	}

	return &Publisher{
		config: config,
	}
}

func NewDefaultPublisher() *Publisher {
	return NewPublisher(PublisherConfig{})
}

func (p *Publisher) Publish(
	ctx context.Context,
	channel *amqp091.Channel,
	exchange, routingKey string,
	mandatory, immediate bool,
	msg amqp091.Publishing,
) error {
	spanName := p.config.SpanNameFormatter(exchange, routingKey)

	spanOpts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindProducer),
	}

	attrs := internal.GetPublishAttributes(exchange, routingKey, &msg)
	for _, attr := range attrs {
		spanOpts = append(spanOpts, trace.WithAttributes(attr))
	}

	if p.config.AttributeEnricher != nil {
		customOpts := p.config.AttributeEnricher(ctx, exchange, routingKey, &msg)
		spanOpts = append(spanOpts, customOpts...)
	}

	ctx, span := p.config.Tracer.Start(ctx, spanName, spanOpts...)
	defer span.End()

	if msg.Headers == nil {
		msg.Headers = make(amqp091.Table)
	}
	p.config.Propagator.InjectToPublishing(ctx, &msg)

	err := channel.Publish(exchange, routingKey, mandatory, immediate, msg)

	internal.SafeSetSpanStatus(span, err)

	return err
}

func (p *Publisher) PublishWithConfirm(
	ctx context.Context,
	channel *amqp091.Channel,
	exchange, routingKey string,
	mandatory, immediate bool,
	msg amqp091.Publishing,
) (*amqp091.DeferredConfirmation, error) {
	spanName := p.config.SpanNameFormatter(exchange, routingKey)

	spanOpts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindProducer),
	}

	attrs := internal.GetPublishAttributes(exchange, routingKey, &msg)
	for _, attr := range attrs {
		spanOpts = append(spanOpts, trace.WithAttributes(attr))
	}

	if p.config.AttributeEnricher != nil {
		customOpts := p.config.AttributeEnricher(ctx, exchange, routingKey, &msg)
		spanOpts = append(spanOpts, customOpts...)
	}

	ctx, span := p.config.Tracer.Start(ctx, spanName, spanOpts...)
	defer span.End()

	if msg.Headers == nil {
		msg.Headers = make(amqp091.Table)
	}
	p.config.Propagator.InjectToPublishing(ctx, &msg)

	confirmation, err := channel.PublishWithDeferredConfirmWithContext(
		ctx, exchange, routingKey, mandatory, immediate, msg,
	)

	internal.SafeSetSpanStatus(span, err)

	return confirmation, err
}

func defaultPublishSpanName(exchange, routingKey string) string {
	if exchange != "" {
		return fmt.Sprintf("%s publish", exchange)
	}
	if routingKey != "" {
		return fmt.Sprintf("%s publish", routingKey)
	}
	return "rabbitmq publish"
}

var defaultPublisher = NewDefaultPublisher()

func Publish(
	ctx context.Context,
	channel *amqp091.Channel,
	exchange, routingKey string,
	mandatory, immediate bool,
	msg amqp091.Publishing,
) error {
	return defaultPublisher.Publish(ctx, channel, exchange, routingKey, mandatory, immediate, msg)
}

func PublishWithConfirm(
	ctx context.Context,
	channel *amqp091.Channel,
	exchange, routingKey string,
	mandatory, immediate bool,
	msg amqp091.Publishing,
) (*amqp091.DeferredConfirmation, error) {
	return defaultPublisher.PublishWithConfirm(ctx, channel, exchange, routingKey, mandatory, immediate, msg)
}
