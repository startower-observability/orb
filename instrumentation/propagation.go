package instrumentation

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	"github.com/startower-observability/orb/internal"
)

type Propagator struct {
}

func NewPropagator() *Propagator {
	return &Propagator{}
}

func (p *Propagator) InjectToPublishing(ctx context.Context, publishing *amqp091.Publishing) {
	if publishing.Headers == nil {
		publishing.Headers = make(amqp091.Table)
	}
	internal.InjectContext(ctx, publishing.Headers)
}

func (p *Propagator) ExtractFromDelivery(ctx context.Context, delivery *amqp091.Delivery) context.Context {
	return internal.ExtractContext(ctx, delivery.Headers)
}

func (p *Propagator) InjectToHeaders(ctx context.Context, headers amqp091.Table) {
	internal.InjectContext(ctx, headers)
}

func (p *Propagator) ExtractFromHeaders(ctx context.Context, headers amqp091.Table) context.Context {
	return internal.ExtractContext(ctx, headers)
}

var DefaultPropagator = NewPropagator()

func InjectToPublishing(ctx context.Context, publishing *amqp091.Publishing) {
	DefaultPropagator.InjectToPublishing(ctx, publishing)
}

func ExtractFromDelivery(ctx context.Context, delivery *amqp091.Delivery) context.Context {
	return DefaultPropagator.ExtractFromDelivery(ctx, delivery)
}
