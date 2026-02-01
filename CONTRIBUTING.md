# Contributing to Factorio Server Manager

Thank you for your interest in contributing to Factorio Server Manager! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Development Setup](#development-setup)
- [Code Style Guidelines](#code-style-guidelines)
- [Pull Request Process](#pull-request-process)
- [Testing Requirements](#testing-requirements)
- [Build Commands](#build-commands)
- [Architecture Overview](#architecture-overview)

## Development Setup

### Prerequisites

- **Go** 1.21 or later
- **Node.js** 16+ and npm
- **Git**
- **Make** (optional, for using Makefile commands)

### Backend Setup (Go)

1. Clone the repository:
   ```bash
   git clone https://github.com/OpenFactorioServerManager/factorio-server-manager.git
   cd factorio-server-manager
   ```

2. Install Go dependencies:
   ```bash
   cd src
   go mod download
   ```

3. Build the backend:
   ```bash
   go build -o ../factorio-server-manager/factorio-server-manager .
   ```

### Frontend Setup (React)

1. Navigate to the UI directory:
   ```bash
   cd ui
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Build for production:
   ```bash
   npm run build
   ```

4. Or run in development mode with hot reload:
   ```bash
   npm run watch
   ```

### Full Build

Use the Makefile for a complete build:
```bash
make build
```

## Code Style Guidelines

### Go Code

- Follow the [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` to format all code
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Handle all errors explicitly - do not ignore them
- Use early returns to reduce nesting
- Prefer composition over inheritance

**Do:**
```go
if err != nil {
    return fmt.Errorf("failed to open file: %w", err)
}
```

**Don't:**
```go
if err != nil {
    panic(err)  // Avoid panics in library code
}

_ = someFunction()  // Don't ignore errors
```

### React/JavaScript Code

- Use functional components with hooks
- Use meaningful component and variable names
- Add PropTypes to all components
- Handle errors in async operations
- Use semantic HTML elements
- Follow React best practices for keys (no array indices)

**Do:**
```jsx
{items.map(item => (
    <ItemComponent key={item.id} item={item} />
))}
```

**Don't:**
```jsx
{items.map((item, index) => (
    <ItemComponent key={index} item={item} />  // Avoid index as key
))}
```

### CSS/Styling

- Use Tailwind CSS utility classes
- Follow BEM naming convention for custom CSS
- Keep styles modular and reusable

## Pull Request Process

1. **Fork the repository** and create your branch from `master`:
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Make your changes** following the code style guidelines

3. **Write or update tests** for any new functionality

4. **Ensure all tests pass**:
   ```bash
   cd src && go test ./... -v
   ```

5. **Ensure the build succeeds**:
   ```bash
   make build
   ```

6. **Commit your changes** with clear, descriptive messages:
   ```bash
   git commit -m "Add feature: description of the feature"
   ```

7. **Push to your fork** and create a Pull Request

### PR Requirements

- Clear title and description explaining the changes
- Reference any related issues
- All tests must pass
- Code must be formatted properly
- New features should include tests
- Documentation updates if needed

### Commit Message Format

- Use present tense ("Add feature" not "Added feature")
- Use imperative mood ("Move cursor to..." not "Moves cursor to...")
- Keep the first line under 72 characters
- Reference issues when applicable: "Fix #123: Description"

## Testing Requirements

### Backend Tests

```bash
cd src
go test ./... -v              # Run all tests
go test ./... -v -test.short  # Skip integration tests
```

For mod-related tests, create `src/.env` with:
```
factorio_username=your_username
factorio_password=your_password
```

### What to Test

- All API handlers should have corresponding tests
- Business logic in the `factorio` package
- Authentication and authorization flows
- Error handling paths

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    // Arrange
    input := "test input"
    expected := "expected output"

    // Act
    result := FunctionUnderTest(input)

    // Assert
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

## Build Commands

| Command | Description |
|---------|-------------|
| `make build` | Build frontend and backend |
| `make clean` | Remove build artifacts |
| `make gen_release` | Build release binaries for Linux and Windows |
| `npm run build` | Build frontend for production |
| `npm run watch` | Frontend development with hot reload |
| `go build ./...` | Build backend |
| `go test ./... -v` | Run all tests |

## Architecture Overview

### Backend (Go) - `/src`

```
src/
├── api/              # REST API handlers and routes
│   ├── handlers.go   # HTTP request handlers
│   ├── routes.go     # API route definitions
│   ├── auth.go       # Authentication middleware
│   └── websocket/    # WebSocket implementation
├── factorio/         # Factorio server control logic
│   ├── server.go     # Server lifecycle management
│   ├── save.go       # Save file operations
│   ├── mods.go       # Mod management
│   └── ...
├── bootstrap/        # Configuration and initialization
└── lockfile/         # File locking utilities
```

### Frontend (React) - `/ui`

```
ui/
├── App/
│   ├── components/   # Reusable UI components
│   └── views/        # Page components
├── api/              # API client modules
├── index.jsx         # Application entry point
└── App.jsx           # Main application component
```

### Key Integration Points

1. **REST API**: Backend serves JSON API at `/api/*`
2. **WebSocket**: Real-time game console at `/ws`
3. **Static Files**: Frontend served from `/app`
4. **RCON**: Internal communication with Factorio server

## Getting Help

- Open an issue for bugs or feature requests
- Check existing issues before creating new ones
- Join discussions in pull requests

## License

By contributing to Factorio Server Manager, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing!
