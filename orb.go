// Package orb provides OpenTelemetry instrumentation for RabbitMQ using amqp091-go.
//
// StarTower Orb enables automatic distributed tracing for RabbitMQ operations
// with minimal code changes. It provides wrappers around amqp091-go types
// that automatically create spans and propagate trace context via message headers.
//
// Basic usage:
//
//	import orb "github.com/startower-observability/orb/instrumentation"
//
//	// Connect with instrumentation
//	conn, err := orb.Dial("amqp://localhost:5672/")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer conn.Close()
//
//	// Create instrumented channel
//	ch, err := conn.ChannelWithTracing()
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer ch.Close()
//
//	// Publish with tracing
//	err = ch.PublishWithTracing(ctx, "exchange", "key", false, false, msg)
//
//	// Consume with tracing
//	handler := func(ctx context.Context, delivery amqp091.Delivery) error {
//		// Process message with trace context
//		return nil
//	}
//	err = ch.ConsumeWithTracing(ctx, "queue", "", false, false, false, false, nil, handler)
//
// The library follows OpenTelemetry semantic conventions and provides
// configurable tracers, propagators, and span attributes.
package orb

import (
	"github.com/startower-observability/orb/instrumentation"
)

type (
	Channel          = instrumentation.Channel
	Connection       = instrumentation.Connection
	Publisher        = instrumentation.Publisher
	Consumer         = instrumentation.Consumer
	Propagator       = instrumentation.Propagator
	MessageHandler   = instrumentation.MessageHandler
	ChannelConfig    = instrumentation.ChannelConfig
	ConnectionConfig = instrumentation.ConnectionConfig
	PublisherConfig  = instrumentation.PublisherConfig
	ConsumerConfig   = instrumentation.ConsumerConfig
)

var (
	Dial                 = instrumentation.Dial
	DialWithConfig       = instrumentation.DialWithConfig
	DialConfig           = instrumentation.DialConfig
	DialConfigWithConfig = instrumentation.DialConfigWithConfig
	NewChannel           = instrumentation.NewChannel
	NewDefaultChannel    = instrumentation.NewDefaultChannel
	NewConnection        = instrumentation.NewConnection
	NewDefaultConnection = instrumentation.NewDefaultConnection
	NewPublisher         = instrumentation.NewPublisher
	NewDefaultPublisher  = instrumentation.NewDefaultPublisher
	NewConsumer          = instrumentation.NewConsumer
	NewDefaultConsumer   = instrumentation.NewDefaultConsumer
	NewPropagator        = instrumentation.NewPropagator
	Publish              = instrumentation.Publish
	PublishWithConfirm   = instrumentation.PublishWithConfirm
	ConsumeWithHandler   = instrumentation.ConsumeWithHandler
	ProcessDelivery      = instrumentation.ProcessDelivery
	WrapDelivery         = instrumentation.WrapDelivery
	InjectToPublishing   = instrumentation.InjectToPublishing
	ExtractFromDelivery  = instrumentation.ExtractFromDelivery
	DefaultPropagator    = instrumentation.DefaultPropagator
)
