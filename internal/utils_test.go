package internal

import (
	"context"
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestHeaderCarrier(t *testing.T) {
	headers := make(amqp091.Table)
	carrier := HeaderCarrier(headers)

	carrier.Set("test-key", "test-value")
	if got := carrier.Get("test-key"); got != "test-value" {
		t.Errorf("Get() = %v, want %v", got, "test-value")
	}

	if got := carrier.Get("non-existent"); got != "" {
		t.Errorf("Get() = %v, want empty string", got)
	}

	carrier.Set("key1", "value1")
	carrier.Set("key2", "value2")
	keys := carrier.Keys()
	if len(keys) != 3 {
		t.Errorf("Keys() length = %v, want 3", len(keys))
	}
}

func TestGetCommonAttributes(t *testing.T) {
	tests := []struct {
		name       string
		exchange   string
		routingKey string
		wantAttrs  []attribute.KeyValue
	}{
		{
			name:       "exchange with routing key",
			exchange:   "test-exchange",
			routingKey: "test.key",
			wantAttrs: []attribute.KeyValue{
				attribute.String(MessagingSystem, SystemRabbitMQ),
				attribute.String(MessagingDestination, "test-exchange"),
				attribute.String(MessagingDestinationKind, DestinationKindTopic),
				attribute.String(MessagingRabbitMQRoutingKey, "test.key"),
			},
		},
		{
			name:       "direct queue",
			exchange:   "",
			routingKey: "test-queue",
			wantAttrs: []attribute.KeyValue{
				attribute.String(MessagingSystem, SystemRabbitMQ),
				attribute.String(MessagingDestinationKind, DestinationKindQueue),
				attribute.String(MessagingDestination, "test-queue"),
				attribute.String(MessagingRabbitMQRoutingKey, "test-queue"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := GetCommonAttributes(tt.exchange, tt.routingKey)

			if len(attrs) != len(tt.wantAttrs) {
				t.Errorf("GetCommonAttributes() length = %v, want %v", len(attrs), len(tt.wantAttrs))
				return
			}

			for i, want := range tt.wantAttrs {
				if attrs[i] != want {
					t.Errorf("GetCommonAttributes()[%d] = %v, want %v", i, attrs[i], want)
				}
			}
		})
	}
}

func TestGetPublishAttributes(t *testing.T) {
	msg := &amqp091.Publishing{
		MessageId:     "msg-123",
		CorrelationId: "corr-456",
	}

	attrs := GetPublishAttributes("test-exchange", "test.key", msg)

	found := make(map[string]string)
	for _, attr := range attrs {
		found[string(attr.Key)] = attr.Value.AsString()
	}

	if found[MessagingSystem] != SystemRabbitMQ {
		t.Errorf("Missing or incorrect messaging.system attribute")
	}

	if found[MessagingOperation] != OperationPublish {
		t.Errorf("Missing or incorrect messaging.operation attribute")
	}

	if found[MessagingMessageID] != "msg-123" {
		t.Errorf("Missing or incorrect messaging.message_id attribute")
	}

	if found[MessagingConversationID] != "corr-456" {
		t.Errorf("Missing or incorrect messaging.conversation_id attribute")
	}
}

func TestGetConsumeAttributes(t *testing.T) {
	delivery := &amqp091.Delivery{
		RoutingKey:    "test.key",
		MessageId:     "msg-123",
		CorrelationId: "corr-456",
	}

	attrs := GetConsumeAttributes("test-queue", delivery)

	found := make(map[string]string)
	for _, attr := range attrs {
		found[string(attr.Key)] = attr.Value.AsString()
	}

	if found[MessagingSystem] != SystemRabbitMQ {
		t.Errorf("Missing or incorrect messaging.system attribute")
	}

	if found[MessagingOperation] != OperationReceive {
		t.Errorf("Missing or incorrect messaging.operation attribute")
	}

	if found[MessagingDestination] != "test-queue" {
		t.Errorf("Missing or incorrect messaging.destination attribute")
	}
}

func TestInjectExtractContext(t *testing.T) {
	ctx := context.Background()

	headers := make(amqp091.Table)

	InjectContext(ctx, headers)

	extractedCtx := ExtractContext(context.Background(), headers)

	if extractedCtx == nil {
		t.Error("ExtractContext returned nil")
	}
}

func TestSafeSetSpanStatus(t *testing.T) {
	var span trace.Span
	SafeSetSpanStatus(span, nil)

	err := &testError{msg: "test error"}
	SafeSetSpanStatus(span, err)
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
