package instrumentation

import (
	"context"
	"testing"

	"github.com/rabbitmq/amqp091-go"
)

func TestNewPropagator(t *testing.T) {
	p1 := NewPropagator()
	if p1 == nil {
		t.Error("NewPropagator() returned nil")
	}
}

func TestPropagatorInjectToPublishing(t *testing.T) {
	p := NewPropagator()
	ctx := context.Background()

	publishing := &amqp091.Publishing{}
	p.InjectToPublishing(ctx, publishing)

	if publishing.Headers == nil {
		t.Error("InjectToPublishing should create Headers if nil")
	}

	publishing.Headers = amqp091.Table{"existing": "value"}
	p.InjectToPublishing(ctx, publishing)

	if publishing.Headers["existing"] != "value" {
		t.Error("InjectToPublishing should preserve existing headers")
	}
}

func TestPropagatorExtractFromDelivery(t *testing.T) {
	p := NewPropagator()
	ctx := context.Background()

	delivery := &amqp091.Delivery{}
	extractedCtx := p.ExtractFromDelivery(ctx, delivery)

	if extractedCtx == nil {
		t.Error("ExtractFromDelivery returned nil context")
	}

	delivery.Headers = amqp091.Table{"test": "value"}
	extractedCtx = p.ExtractFromDelivery(ctx, delivery)

	if extractedCtx == nil {
		t.Error("ExtractFromDelivery returned nil context with headers")
	}
}

func TestDefaultPropagator(t *testing.T) {
	if DefaultPropagator == nil {
		t.Error("DefaultPropagator is nil")
	}
}

func TestConvenienceFunctions(t *testing.T) {
	ctx := context.Background()

	publishing := &amqp091.Publishing{}
	InjectToPublishing(ctx, publishing)

	if publishing.Headers == nil {
		t.Error("InjectToPublishing should create Headers")
	}

	delivery := &amqp091.Delivery{
		Headers: amqp091.Table{"test": "value"},
	}
	extractedCtx := ExtractFromDelivery(ctx, delivery)

	if extractedCtx == nil {
		t.Error("ExtractFromDelivery returned nil context")
	}
}
