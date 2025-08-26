package internal

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	TracerName                  = "github.com/startower-observability/orb"
	MessagingSystem             = "messaging.system"
	MessagingDestinationKind    = "messaging.destination_kind"
	MessagingDestination        = "messaging.destination"
	MessagingRabbitMQRoutingKey = "messaging.rabbitmq.routing_key"
	MessagingOperation          = "messaging.operation"
	MessagingMessageID          = "messaging.message_id"
	MessagingConversationID     = "messaging.conversation_id"
	SystemRabbitMQ              = "rabbitmq"
	DestinationKindQueue        = "queue"
	DestinationKindTopic        = "topic"
	OperationPublish            = "publish"
	OperationReceive            = "receive"
	OperationProcess            = "process"
)

type HeaderCarrier amqp091.Table

func (hc HeaderCarrier) Get(key string) string {
	if val, ok := hc[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (hc HeaderCarrier) Set(key, value string) {
	hc[key] = value
}

func (hc HeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(hc))
	for k := range hc {
		keys = append(keys, k)
	}
	return keys
}

func GetCommonAttributes(exchange, routingKey string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String(MessagingSystem, SystemRabbitMQ),
	}

	if exchange != "" {
		attrs = append(attrs, attribute.String(MessagingDestination, exchange))
		attrs = append(attrs, attribute.String(MessagingDestinationKind, DestinationKindTopic))
	} else {
		attrs = append(attrs, attribute.String(MessagingDestinationKind, DestinationKindQueue))
		if routingKey != "" {
			attrs = append(attrs, attribute.String(MessagingDestination, routingKey))
		}
	}

	if routingKey != "" {
		attrs = append(attrs, attribute.String(MessagingRabbitMQRoutingKey, routingKey))
	}

	return attrs
}

func GetPublishAttributes(exchange, routingKey string, msg *amqp091.Publishing) []attribute.KeyValue {
	attrs := GetCommonAttributes(exchange, routingKey)
	attrs = append(attrs, attribute.String(MessagingOperation, OperationPublish))

	if msg.MessageId != "" {
		attrs = append(attrs, attribute.String(MessagingMessageID, msg.MessageId))
	}

	if msg.CorrelationId != "" {
		attrs = append(attrs, attribute.String(MessagingConversationID, msg.CorrelationId))
	}

	return attrs
}

func GetConsumeAttributes(queueName string, delivery *amqp091.Delivery) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String(MessagingSystem, SystemRabbitMQ),
		attribute.String(MessagingDestinationKind, DestinationKindQueue),
		attribute.String(MessagingOperation, OperationReceive),
	}

	if queueName != "" {
		attrs = append(attrs, attribute.String(MessagingDestination, queueName))
	}

	if delivery.RoutingKey != "" {
		attrs = append(attrs, attribute.String(MessagingRabbitMQRoutingKey, delivery.RoutingKey))
	}

	if delivery.MessageId != "" {
		attrs = append(attrs, attribute.String(MessagingMessageID, delivery.MessageId))
	}

	if delivery.CorrelationId != "" {
		attrs = append(attrs, attribute.String(MessagingConversationID, delivery.CorrelationId))
	}

	return attrs
}

func InjectContext(ctx context.Context, headers amqp091.Table) {
	if headers == nil {
		return
	}
	otel.GetTextMapPropagator().Inject(ctx, HeaderCarrier(headers))
}

func ExtractContext(ctx context.Context, headers amqp091.Table) context.Context {
	if headers == nil {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, HeaderCarrier(headers))
}

func SafeSetSpanStatus(span trace.Span, err error) {
	if span == nil {
		return
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
}
