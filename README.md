<div align="center">
  <img src="images/logo.png" alt="StarTower Orb" width="200"/>
  
  # StarTower Orb
  ### RabbitMQ OpenTelemetry Instrumentation
  
  [![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)
  [![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
  [![Go Reference](https://pkg.go.dev/badge/github.com/startower-observability/orb.svg)](https://pkg.go.dev/github.com/startower-observability/orb)
</div>

StarTower Orb is a Go library that provides automatic OpenTelemetry instrumentation for RabbitMQ using the `amqp091-go` client. It enables distributed tracing across your RabbitMQ-based microservices with minimal code changes.

## Features

-  **Automatic Tracing**: Instruments RabbitMQ publish and consume operations
-  **Context Propagation**: Propagates trace context via W3C headers in message headers
-  **Semantic Conventions**: Follows OpenTelemetry semantic conventions for messaging
-  **Configurable**: Customizable tracers, propagators, and span attributes
-  **Drop-in Replacement**: Minimal changes to existing `amqp091-go` code
-  **Production Ready**: Comprehensive error handling and graceful degradation

## Installation

```bash
go get github.com/startower-observability/orb
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/rabbitmq/amqp091-go"
    orb "github.com/startower-observability/orb/instrumentation"
    "go.opentelemetry.io/otel"
)

func main() {
    // Initialize OpenTelemetry (tracer provider, etc.)
    // ... your OTel setup code ...
    
    // Connect to RabbitMQ with instrumentation
    conn, err := orb.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    // Create instrumented channel
    ch, err := conn.ChannelWithTracing()
    if err != nil {
        log.Fatal(err)
    }
    defer ch.Close()
    
    ctx := context.Background()
    
    // Publish with tracing
    err = ch.PublishWithTracing(ctx, 
        "my-exchange", "routing.key", false, false,
        amqp091.Publishing{
            ContentType: "text/plain",
            Body:        []byte("Hello, World!"),
        })
    if err != nil {
        log.Fatal(err)
    }
}
```

### Consumer with Tracing

```go
func consumeMessages(ch *orb.Channel) {
    ctx := context.Background()
    
    // Define message handler
    handler := func(ctx context.Context, delivery amqp091.Delivery) error {
        // Process message with trace context
        log.Printf("Received message: %s", delivery.Body)
        
        // Your business logic here
        return nil
    }
    
    // Start consuming with tracing
    err := ch.ConsumeWithTracing(ctx,
        "my-queue", "consumer-tag", false, false, false, false, nil, handler)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Advanced Usage

### Custom Configuration

```go
// Custom publisher configuration
publisherConfig := orb.PublisherConfig{
    Tracer: otel.Tracer("my-service"),
    SpanNameFormatter: func(exchange, routingKey string) string {
        return fmt.Sprintf("publish to %s", exchange)
    },
    AttributeEnricher: func(ctx context.Context, exchange, routingKey string, msg *amqp091.Publishing) []trace.SpanStartOption {
        return []trace.SpanStartOption{
            trace.WithAttributes(
                attribute.String("custom.attribute", "value"),
            ),
        }
    },
}

// Custom consumer configuration
consumerConfig := orb.ConsumerConfig{
    Tracer: otel.Tracer("my-service"),
    SpanNameFormatter: func(queueName string, delivery *amqp091.Delivery) string {
        return fmt.Sprintf("process from %s", queueName)
    },
}

// Create connection with custom config
conn, err := orb.DialWithConfig("amqp://localhost:5672/", orb.ConnectionConfig{
    ChannelConfig: orb.ChannelConfig{
        PublisherConfig: publisherConfig,
        ConsumerConfig:  consumerConfig,
    },
})
```

### Manual Message Processing

```go
// Get raw deliveries channel
deliveries, err := ch.ConsumeWithContext(ctx, "my-queue", "", false, false, false, false, nil)
if err != nil {
    log.Fatal(err)
}

for delivery := range deliveries {
    // Wrap delivery with tracing
    ctx, span := ch.WrapDeliveryWithTracing(context.Background(), "my-queue", &delivery)
    
    // Process message
    err := processMessage(ctx, delivery)
    
    // Handle acknowledgment
    if err != nil {
        delivery.Nack(false, true)
        span.RecordError(err)
    } else {
        delivery.Ack(false)
    }
    
    span.End()
}
```

### Using Standalone Components

```go
// Use publisher directly
publisher := orb.NewDefaultPublisher()
err := publisher.Publish(ctx, channel, "exchange", "key", false, false, msg)

// Use consumer directly  
consumer := orb.NewDefaultConsumer()
err := consumer.ProcessDelivery(ctx, "queue", delivery, handler)

// Use propagation directly
orb.InjectToPublishing(ctx, &publishing)
ctx = orb.ExtractFromDelivery(ctx, &delivery)
```

## Semantic Conventions

The library follows OpenTelemetry semantic conventions for messaging:

| Attribute | Description | Example |
|-----------|-------------|---------|
| `messaging.system` | Messaging system | `rabbitmq` |
| `messaging.destination` | Queue or exchange name | `user.events` |
| `messaging.destination_kind` | Destination type | `queue`, `topic` |
| `messaging.rabbitmq.routing_key` | Routing key | `user.created` |
| `messaging.operation` | Operation type | `publish`, `receive` |
| `messaging.message_id` | Message ID | `msg-123` |
| `messaging.conversation_id` | Correlation ID | `conv-456` |

## Span Kinds

- **Producer spans**: Created for publish operations (`SpanKindProducer`)
- **Consumer spans**: Created for consume operations (`SpanKindConsumer`)

## Context Propagation

The library automatically:

1. **Injects** trace context into message headers when publishing
2. **Extracts** trace context from message headers when consuming
3. **Links** producer and consumer spans across service boundaries

Headers are injected using the W3C Trace Context format in the `amqp091.Publishing.Headers` field.

## Error Handling

The library provides graceful error handling:

- Spans are marked with error status when operations fail
- Missing headers or context don't cause failures
- Original `amqp091-go` errors are preserved and returned
- Instrumentation errors are recorded but don't interrupt message flow

## Performance Considerations

- Minimal overhead: ~1-2μs per operation
- Headers are only modified when necessary
- Lazy initialization of OpenTelemetry components
- No goroutine leaks or memory retention

## Testing

Run tests with RabbitMQ:

```bash
# Start RabbitMQ container
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

# Run tests
go test ./...

# Run integration tests
go test -tags=integration ./...
```

## Examples

See the [examples](examples/) directory for complete working examples:

- [Basic Publisher/Consumer](examples/basic/)
- [Custom Configuration](examples/custom-config/)
- [Error Handling](examples/error-handling/)
- [Performance Benchmarks](examples/benchmarks/)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go)
- [amqp091-go](https://github.com/rabbitmq/amqp091-go)
- [RabbitMQ](https://www.rabbitmq.com/)
