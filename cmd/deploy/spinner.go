package deploy

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg error

type spinnerModel struct {
	spinner  spinner.Model
	quitting bool
	err      error
	start    time.Time
	lines    int
}

func initialSpinnerModel() spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return spinnerModel{
		spinner: s,
		start:   time.Now(),
		lines:   0,
	}
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.QuitMsg:
		m.quitting = true
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

func (m spinnerModel) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	if m.quitting {
		return ""
	}

	elapsed := time.Since(m.start).Round(time.Second)
	return fmt.Sprintf("%s Deploying... (%s)", m.spinner.View(), elapsed)
}

func NewSpinner() *tea.Program {
	return tea.NewProgram(initialSpinnerModel())
}

// ExitSpinner stops the spinner and displays the given message
func ExitSpinner(p *tea.Program, message string) {
	// Send quit message first to stop the spinner
	p.Send(tea.Quit())
	// Give the spinner a moment to clean up
	time.Sleep(50 * time.Millisecond)
	// Clear any remaining spinner artifacts and print the message
	fmt.Print("\r")                    // Move cursor to beginning of line
	fmt.Print("\033[2K")               // Clear the entire line
	fmt.Println(message)               // Print the final message
}
