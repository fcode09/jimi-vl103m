# Contributing to Jimi VL103M Decoder

Thank you for your interest in contributing to the Jimi VL103M Protocol Decoder! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and collaborative environment.

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/intelcon-group/jimi-vl103m/issues)
2. If not, create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - Hex dump of problematic packet (if applicable)
   - Go version and OS

### Suggesting Enhancements

1. Check existing issues and discussions
2. Create a new issue with:
   - Clear use case description
   - Proposed solution (if any)
   - Examples

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`make test`)
6. Run linter (`make lint`)
7. Commit with clear messages
8. Push to your fork
9. Create a Pull Request

## Development Guidelines

### Code Style

- Follow standard Go conventions
- Use `gofmt` and `goimports`
- Keep functions focused and small
- Add comments for exported types and functions
- Use meaningful variable names

### Testing

- Add unit tests for new code
- Maintain test coverage >90%
- Include table-driven tests where appropriate
- Add integration tests for complex features
- Test with real packet data when possible

### Documentation

- Update README.md if adding features
- Add godoc comments for exported APIs
- Update CHANGELOG.md
- Include examples for new functionality

### Commit Messages

Follow conventional commits format:

```
type(scope): subject

body (optional)

footer (optional)
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:
- `feat(decoder): add support for 0x94 info transfer packet`
- `fix(parser): correct latitude parsing for southern hemisphere`
- `docs(readme): add kafka integration example`

## Project Structure

```
pkg/          # Public API (exported)
internal/     # Internal implementation (not exported)
cmd/          # Command-line tools
examples/     # Usage examples
test/         # Integration tests
```

## Testing Locally

```bash
# Run all tests
make test

# Run with race detector
go test -race ./...

# Run with coverage
make test-coverage

# Run benchmarks
make benchmark

# Run linter
make lint
```

## Questions?

Feel free to open an issue for discussion or contact the maintainers.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
