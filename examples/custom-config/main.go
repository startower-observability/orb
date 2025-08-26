package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	orb "github.com/startower-observability/orb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func main() {
	// Initialize OpenTelemetry
	initTracer()

	// Create custom configuration
	config := orb.ConnectionConfig{
		ChannelConfig: orb.ChannelConfig{
			PublisherConfig: orb.PublisherConfig{
				Tracer: otel.Tracer("my-service-publisher"),
				SpanNameFormatter: func(exchange, routingKey string) string {
					return fmt.Sprintf("📤 Publish to %s/%s", exchange, routingKey)
				},
				AttributeEnricher: func(ctx context.Context, exchange, routingKey string, msg *amqp091.Publishing) []oteltrace.SpanStartOption {
					return []oteltrace.SpanStartOption{
						oteltrace.WithAttributes(
							attribute.String("service.name", "my-service"),
							attribute.String("message.type", "order"),
							attribute.Int("message.size", len(msg.Body)),
						),
					}
				},
			},
			ConsumerConfig: orb.ConsumerConfig{
				Tracer: otel.Tracer("my-service-consumer"),
				SpanNameFormatter: func(queueName string, delivery *amqp091.Delivery) string {
					return fmt.Sprintf("📥 Process from %s", queueName)
				},
				AttributeEnricher: func(ctx context.Context, queueName string, delivery *amqp091.Delivery) []oteltrace.SpanStartOption {
					return []oteltrace.SpanStartOption{
						oteltrace.WithAttributes(
							attribute.String("service.name", "my-service"),
							attribute.String("queue.name", queueName),
							attribute.Int("message.size", len(delivery.Body)),
						),
					}
				},
			},
		},
	}

	// Connect to RabbitMQ with custom configuration
	conn, err := orb.DialWithConfig("amqp://guest:guest@localhost:5672/", config)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Create instrumented channel
	ch, err := conn.ChannelWithTracing()
	if err != nil {
		log.Fatalf("Failed to create channel: %v", err)
	}
	defer ch.Close()

	// Declare exchange and queue
	err = ch.ExchangeDeclare(
		"orders", // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	queue, err := ch.QueueDeclare(
		"order-processing", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	err = ch.QueueBind(
		queue.Name,      // queue name
		"order.created", // routing key
		"orders",        // exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind queue: %v", err)
	}

	ctx := context.Background()

	// Publish multiple messages with custom tracing
	log.Println("Publishing orders...")
	for i := 1; i <= 3; i++ {
		orderData := fmt.Sprintf(`{"orderId": %d, "amount": %d.99}`, i, i*10)

		err = ch.PublishWithTracing(ctx,
			"orders",        // exchange
			"order.created", // routing key
			false,           // mandatory
			false,           // immediate
			amqp091.Publishing{
				ContentType:   "application/json",
				Body:          []byte(orderData),
				MessageId:     fmt.Sprintf("order-%d", i),
				CorrelationId: fmt.Sprintf("corr-%d", i),
				Timestamp:     time.Now(),
			})
		if err != nil {
			log.Fatalf("Failed to publish order %d: %v", i, err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Start consuming with custom tracing
	log.Println("Starting order processor...")
	handler := func(ctx context.Context, delivery amqp091.Delivery) error {
		log.Printf("Processing order: %s", delivery.Body)

		// Simulate order processing
		time.Sleep(200 * time.Millisecond)

		// Add custom span attributes
		span := oteltrace.SpanFromContext(ctx)
		span.SetAttributes(
			attribute.String("order.status", "processed"),
			attribute.Bool("order.valid", true),
		)

		return nil
	}

	err = ch.ConsumeWithTracing(ctx,
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack (manual ack for better tracing)
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
		handler,
	)
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	// Wait for processing to complete
	time.Sleep(3 * time.Second)
	log.Println("Custom configuration example completed!")
}

func initTracer() {
	// Create stdout exporter with pretty printing
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		log.Fatalf("Failed to create stdout exporter: %v", err)
	}

	// Create tracer provider with custom resource
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	log.Println("OpenTelemetry tracer initialized with custom configuration")
}
