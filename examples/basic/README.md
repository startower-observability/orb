# Basic Example

This example demonstrates basic usage of StarTower Orb for RabbitMQ instrumentation.

## Prerequisites

1. RabbitMQ server running on localhost:5672
2. Go 1.24+

## Running RabbitMQ with Docker

```bash
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management
```

## Running the Example

```bash
cd examples/basic
go mod tidy
go run main.go
```

## What it does

1. Initializes OpenTelemetry with stdout exporter
2. Connects to RabbitMQ with instrumentation
3. Declares a queue
4. Publishes a message with automatic tracing
5. Consumes the message with automatic tracing
6. Shows trace output in stdout

## Expected Output

You should see trace spans for both publish and consume operations, showing the distributed tracing in action.
