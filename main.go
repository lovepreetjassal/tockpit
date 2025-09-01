package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg error

type httpResponse struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       string
	Duration   time.Duration
}

type requestCompleteMsg struct {
	response httpResponse
	err      error
}

type state int

const (
	inputState state = iota
	loadingState
	responseState
)

type model struct {
	state        state
	urlInput     textinput.Model
	methodInput  textinput.Model
	headersInput textarea.Model
	bodyInput    textarea.Model
	spinner      spinner.Model
	response     httpResponse
	err          error
	focused      int // 0: url, 1: method, 2: headers, 3: body
	quitting     bool
}

var (
	quitKeys = key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("", "press q to quit"),
	)

	sendKeys = key.NewBinding(
		key.WithKeys("enter", "ctrl+s"),
		key.WithHelp("", "enter/ctrl+s to send request"),
	)

	tabKeys = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("", "tab to switch fields"),
	)

	backKeys = key.NewBinding(
		key.WithKeys("ctrl+b"),
		key.WithHelp("", "ctrl+b to go back"),
	)
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	focusedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("205"))

	responseStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("32")).
			Padding(1)
)

func makeHttpRequest(url, method string, headers map[string]string, body string) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()

		var reqBody io.Reader
		if body != "" {
			reqBody = strings.NewReader(body)
		}

		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			return requestCompleteMsg{err: err}
		}

		// Set headers
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return requestCompleteMsg{err: err}
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return requestCompleteMsg{err: err}
		}

		duration := time.Since(start)

		return requestCompleteMsg{
			response: httpResponse{
				StatusCode: resp.StatusCode,
				Status:     resp.Status,
				Headers:    resp.Header,
				Body:       string(respBody),
				Duration:   duration,
			},
		}
	}
}

func parseHeaders(headersText string) map[string]string {
	headers := make(map[string]string)
	lines := strings.Split(headersText, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers[key] = value
		}
	}

	return headers
}

func initialModel() model {
	// URL input
	urlInput := textinput.New()
	urlInput.Placeholder = "Enter URL here..."
	urlInput.SetValue("https://httpbin.org/get")
	urlInput.CharLimit = 500
	urlInput.Width = 80

	// Method input
	methodInput := textinput.New()
	methodInput.Placeholder = "GET"
	methodInput.SetValue("GET")
	methodInput.CharLimit = 10
	methodInput.Width = 20

	// Headers input
	headersInput := textarea.New()
	headersInput.Placeholder = "Content-Type: application/json"
	headersInput.SetWidth(80)
	headersInput.SetHeight(4)

	// Body input
	bodyInput := textarea.New()
	bodyInput.Placeholder = "Request body (for POST/PUT)"
	bodyInput.SetWidth(80)
	bodyInput.SetHeight(6)

	// Spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := model{
		state:        inputState,
		urlInput:     urlInput,
		methodInput:  methodInput,
		headersInput: headersInput,
		bodyInput:    bodyInput,
		spinner:      s,
		focused:      0,
	}

	// Set initial focus
	m.urlInput.Focus()
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit
		}

		switch m.state {
		case inputState:
			switch {
			case key.Matches(msg, sendKeys):
				if m.urlInput.Value() == "" {
					return m, nil
				}

				m.state = loadingState
				headers := parseHeaders(m.headersInput.Value())
				method := m.methodInput.Value()
				if method == "" {
					method = "GET"
				}

				return m, tea.Batch(
					m.spinner.Tick,
					makeHttpRequest(m.urlInput.Value(), method, headers, m.bodyInput.Value()),
				)

			case key.Matches(msg, tabKeys):
				m.focused = (m.focused + 1) % 4
				m.updateFocus()

			case msg.String() == "shift+tab":
				m.focused = (m.focused - 1 + 4) % 4
				m.updateFocus()
			}

		case responseState:
			if key.Matches(msg, backKeys) {
				m.state = inputState
				m.updateFocus()
			}
		}

	case requestCompleteMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.response = msg.response
			m.err = nil
		}
		m.state = responseState
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil
	}

	// Update components based on current state
	switch m.state {
	case inputState:
		switch m.focused {
		case 0:
			m.urlInput, cmd = m.urlInput.Update(msg)
			cmds = append(cmds, cmd)
		case 1:
			m.methodInput, cmd = m.methodInput.Update(msg)
			cmds = append(cmds, cmd)
		case 2:
			m.headersInput, cmd = m.headersInput.Update(msg)
			cmds = append(cmds, cmd)
		case 3:
			m.bodyInput, cmd = m.bodyInput.Update(msg)
			cmds = append(cmds, cmd)
		}

	case loadingState:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) updateFocus() {
	m.urlInput.Blur()
	m.methodInput.Blur()
	m.headersInput.Blur()
	m.bodyInput.Blur()

	switch m.focused {
	case 0:
		m.urlInput.Focus()
	case 1:
		m.methodInput.Focus()
	case 2:
		m.headersInput.Focus()
	case 3:
		m.bodyInput.Focus()
	}
}

func (m model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	switch m.state {
	case inputState:
		return m.inputView()
	case loadingState:
		return m.loadingView()
	case responseState:
		return m.responseView()
	default:
		return "Unknown state"
	}
}

func (m model) inputView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🌐 HTTP Endpoint Tester"))
	b.WriteString("\n\n")

	// URL input
	urlStyle := inputStyle
	if m.focused == 0 {
		urlStyle = focusedInputStyle
	}
	b.WriteString("URL:\n")
	b.WriteString(urlStyle.Render(m.urlInput.View()))
	b.WriteString("\n\n")

	// Method input
	methodStyle := inputStyle
	if m.focused == 1 {
		methodStyle = focusedInputStyle
	}
	b.WriteString("Method:\n")
	b.WriteString(methodStyle.Render(m.methodInput.View()))
	b.WriteString("\n\n")

	// Headers input
	headersStyle := inputStyle
	if m.focused == 2 {
		headersStyle = focusedInputStyle
	}
	b.WriteString("Headers (one per line, format: Key: Value):\n")
	b.WriteString(headersStyle.Render(m.headersInput.View()))
	b.WriteString("\n\n")

	// Body input
	bodyStyle := inputStyle
	if m.focused == 3 {
		bodyStyle = focusedInputStyle
	}
	b.WriteString("Request Body:\n")
	b.WriteString(bodyStyle.Render(m.bodyInput.View()))
	b.WriteString("\n\n")

	// Help text
	b.WriteString("📝 Tab/Shift+Tab: switch fields | Enter/Ctrl+S: send request | Q: quit")

	return b.String()
}

func (m model) loadingView() string {
	return fmt.Sprintf("\n\n   %s Making HTTP request...\n\n", m.spinner.View())
}

func (m model) responseView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📡 HTTP Response"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString("❌ Error: ")
		b.WriteString(m.err.Error())
		b.WriteString("\n\n")
	} else {
		// Status
		statusColor := "32" // green
		if m.response.StatusCode >= 400 {
			statusColor = "196" // red
		} else if m.response.StatusCode >= 300 {
			statusColor = "214" // yellow
		}

		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Bold(true)
		b.WriteString("Status: ")
		b.WriteString(statusStyle.Render(m.response.Status))
		b.WriteString(fmt.Sprintf(" (%s)", m.response.Duration.String()))
		b.WriteString("\n\n")

		// Headers
		b.WriteString("Response Headers:\n")
		headersJson, _ := json.MarshalIndent(m.response.Headers, "", "  ")
		b.WriteString(responseStyle.Render(string(headersJson)))
		b.WriteString("\n\n")

		// Body
		b.WriteString("Response Body:\n")
		body := m.response.Body
		if len(body) > 1000 {
			body = body[:1000] + "..."
		}

		// Try to pretty print JSON
		var prettyBody string
		var jsonData interface{}
		if err := json.Unmarshal([]byte(m.response.Body), &jsonData); err == nil {
			if pretty, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
				prettyBody = string(pretty)
				if len(prettyBody) > 1000 {
					prettyBody = prettyBody[:1000] + "..."
				}
			} else {
				prettyBody = body
			}
		} else {
			prettyBody = body
		}

		b.WriteString(responseStyle.Render(prettyBody))
		b.WriteString("\n\n")
	}

	b.WriteString("🔙 Ctrl+B: go back | Q: quit")

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
