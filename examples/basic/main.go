package main

import (
	"context"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	orb "github.com/startower-observability/orb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	// Initialize OpenTelemetry
	initTracer()

	// Connect to RabbitMQ with instrumentation
	conn, err := orb.Dial("amqp://guest:guest@localhost:5672/")
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

	// Declare a queue
	queue, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	ctx := context.Background()

	// Publish a message with tracing
	log.Println("Publishing message...")
	err = ch.PublishWithTracing(ctx,
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte("Hello, World!"),
			MessageId:   "msg-123",
		})
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}

	// Start consuming messages with tracing
	log.Println("Starting consumer...")
	handler := func(ctx context.Context, delivery amqp091.Delivery) error {
		log.Printf("Received message: %s", delivery.Body)
		// Simulate processing time
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	err = ch.ConsumeWithTracing(ctx,
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
		handler,
	)
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	// Wait for a bit to see the message processing
	time.Sleep(2 * time.Second)
	log.Println("Example completed!")
}

func initTracer() {
	// Create stdout exporter for demonstration
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatalf("Failed to create stdout exporter: %v", err)
	}

	// Create tracer provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	log.Println("OpenTelemetry tracer initialized")
}
