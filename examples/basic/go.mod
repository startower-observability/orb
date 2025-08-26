module github.com/startower-observability/orb/examples/basic

go 1.24

require (
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/startower-observability/orb v0.1.0
	go.opentelemetry.io/otel v1.32.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.32.0
	go.opentelemetry.io/otel/sdk v1.32.0
)

replace github.com/startower-observability/orb => ../../
