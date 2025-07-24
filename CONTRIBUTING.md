# Contributing to Parsec

Thank you for your interest in contributing to Parsec! We welcome contributions from everyone.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/your-username/parsec.git`
3. Create a branch: `git checkout -b your-feature-name`
4. Make your changes
5. Test thoroughly
6. Commit your changes: `git commit -m "Description of changes"`
7. Push to your fork: `git push origin your-feature-name`
8. Open a Pull Request

## Development Setup

1. Ensure you have Go 1.24.4 or later installed
2. Install dependencies: `go mod tidy`
3. Build the project: `go build .`

## Code Style

- Follow standard Go formatting guidelines
- Use `gofmt` to format your code
- Write descriptive commit messages
- Add comments for non-obvious code
- Update documentation when needed

## Testing

- Test your changes with various file types
- Verify terminal resizing behavior
- Check performance with large directories
- Test on both Windows and Unix systems

## Pull Request Process

1. Update the README.md with details of significant changes
2. Update the version numbers following [SemVer](http://semver.org/)
3. Ensure all tests pass
4. Link any relevant issues
5. The PR will be merged once you have the sign-off of a maintainer

## Bug Reports

When filing an issue, please include:

- Your operating system and version
- Steps to reproduce the issue
- Expected vs actual behavior
- Any relevant error messages
- Screenshots if applicable

## Feature Requests

We welcome feature requests! Please provide:

- Clear description of the feature
- Use cases and benefits
- Potential implementation approach
- Any relevant examples

## Questions?

Feel free to open an issue for any questions about contributing.

Thank you for helping make Parsec better!