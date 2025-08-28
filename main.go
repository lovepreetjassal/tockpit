package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg error

// Screen represents different views in the application
type Screen int

const (
	HomeScreen Screen = iota
	AboutScreen
	HelpScreen
)

type model struct {
	spinner       spinner.Model
	quitting      bool
	err           error
	currentScreen Screen
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

var navigationKeys = key.NewBinding(
	key.WithKeys("1", "2", "3"),
	key.WithHelp("1-3", "navigate screens"),
)

var helpKeys = key.NewBinding(
	key.WithKeys("h", "?"),
	key.WithHelp("h/?", "show help"),
)

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{
		spinner:       s,
		currentScreen: HomeScreen,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit
		}

		// Handle navigation
		switch msg.String() {
		case "1":
			m.currentScreen = HomeScreen
		case "2":
			m.currentScreen = AboutScreen
		case "3", "h", "?":
			m.currentScreen = HelpScreen
		}

		return m, nil
	case errMsg:
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	if m.quitting {
		return "\nGoodbye!\n\n"
	}

	// Render the appropriate screen
	switch m.currentScreen {
	case HomeScreen:
		return m.homeView()
	case AboutScreen:
		return m.aboutView()
	case HelpScreen:
		return m.helpView()
	default:
		return m.homeView()
	}
}

func (m model) homeView() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("🏠 Tockpit - Home"))
	s.WriteString("\n\n")
	s.WriteString(fmt.Sprintf("   %s Welcome to Tockpit!", m.spinner.View()))
	s.WriteString("\n\n")
	s.WriteString("   This is a terminal-based application built with Bubbletea.")
	s.WriteString("\n\n")
	s.WriteString(m.navigationFooter())
	return s.String()
}

func (m model) aboutView() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("ℹ️  Tockpit - About"))
	s.WriteString("\n\n")
	s.WriteString("   Tockpit is a terminal UI application template.")
	s.WriteString("\n")
	s.WriteString("   Built with:")
	s.WriteString("\n")
	s.WriteString("   • Bubbletea - TUI framework")
	s.WriteString("\n")
	s.WriteString("   • Bubbles - TUI components")
	s.WriteString("\n")
	s.WriteString("   • Lipgloss - Style definitions")
	s.WriteString("\n\n")
	s.WriteString(m.navigationFooter())
	return s.String()
}

func (m model) helpView() string {
	var s strings.Builder
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("❓ Tockpit - Help"))
	s.WriteString("\n\n")
	s.WriteString("   Navigation:")
	s.WriteString("\n")
	s.WriteString("   • Press 1 - Go to Home screen")
	s.WriteString("\n")
	s.WriteString("   • Press 2 - Go to About screen")
	s.WriteString("\n")
	s.WriteString("   • Press 3, h, or ? - Show this help")
	s.WriteString("\n")
	s.WriteString("   • Press q, ESC, or Ctrl+C - Quit")
	s.WriteString("\n\n")
	s.WriteString(m.navigationFooter())
	return s.String()
}

func (m model) navigationFooter() string {
	currentScreen := ""
	switch m.currentScreen {
	case HomeScreen:
		currentScreen = "[1] Home"
	case AboutScreen:
		currentScreen = "[2] About"
	case HelpScreen:
		currentScreen = "[3] Help"
	}

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Navigation: 1=Home 2=About 3=Help | Current: " + currentScreen + " | q=Quit")

	return "   " + footer + "\n"
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
