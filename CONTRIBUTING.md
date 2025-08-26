# Contributing to StarTower Orb

Thank you for your interest in contributing to StarTower Orb! This document provides guidelines for contributing to the project.

## Development Setup

1. **Prerequisites**
   - Go 1.24 or later
   - RabbitMQ server (for testing)
   - Git

2. **Clone the repository**
   ```bash
   git clone https://github.com/startower-observability/orb.git
   cd orb
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Run tests**
   ```bash
   go test ./...
   ```

## Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` to format your code
- Run `go vet` to check for common errors
- Add comments to all exported functions and types
- Keep functions small and focused

## Testing

- Write unit tests for all new functionality
- Ensure all tests pass before submitting a PR
- Add integration tests for complex features
- Use table-driven tests where appropriate

### Running Tests with RabbitMQ

For integration tests, you'll need a running RabbitMQ instance:

```bash
# Using Docker
docker run -d --name rabbitmq-test -p 5672:5672 rabbitmq:3

# Run tests
go test -tags=integration ./...

# Cleanup
docker stop rabbitmq-test && docker rm rabbitmq-test
```

## Pull Request Process

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following the code style guidelines
3. **Add tests** for your changes
4. **Update documentation** if needed
5. **Ensure all tests pass**
6. **Submit a pull request**

### PR Guidelines

- Use a clear and descriptive title
- Describe what your changes do and why
- Reference any related issues
- Keep PRs focused on a single feature or fix
- Update the README if you're adding new functionality

## Reporting Issues

When reporting issues, please include:

- Go version
- RabbitMQ version
- Operating system
- Steps to reproduce the issue
- Expected vs actual behavior
- Any relevant logs or error messages

## Feature Requests

We welcome feature requests! Please:

- Check if the feature already exists or is planned
- Describe the use case and benefits
- Provide examples of how it would be used
- Consider contributing the implementation

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Maintain a professional tone

## Questions?

If you have questions about contributing, feel free to:

- Open an issue for discussion
- Check existing issues and PRs
- Review the documentation

Thank you for contributing to StarTower Orb!
