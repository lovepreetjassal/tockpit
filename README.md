# tockpit

An HTTP endpoint testing tool built with Bubbletea.

## Features

- Interactive terminal UI for testing HTTP endpoints
- Support for all common HTTP methods (GET, POST, PUT, DELETE, etc.)
- Custom headers configuration
- Request body input for POST/PUT requests
- Real-time response display with status codes, headers, and body
- Pretty-printed JSON responses
- Easy navigation between input form and response views

## Usage

```bash
go build
./bubbletea-app-template
```

### Controls

- **Tab/Shift+Tab**: Navigate between input fields
- **Enter/Ctrl+S**: Send HTTP request
- **Ctrl+B**: Go back from response view to input form
- **Q/Esc/Ctrl+C**: Quit application

### Input Fields

1. **URL**: The endpoint you want to test
2. **Method**: HTTP method (GET, POST, PUT, DELETE, etc.)
3. **Headers**: One header per line in format `Key: Value`
4. **Request Body**: JSON or text content for POST/PUT requests

## Technical Details

Built with:
- [Bubbletea][bubbletea] - TUI framework
- [Bubbles][bubbles] - UI components
- [Lipgloss][lipgloss] - Styling and layout
- Go's standard `net/http` package for HTTP requests

[bubbletea]: https://github.com/charmbracelet/bubbletea
[bubbles]: https://github.com/charmbracelet/bubbles
[lipgloss]: https://github.com/charmbracelet/lipgloss
