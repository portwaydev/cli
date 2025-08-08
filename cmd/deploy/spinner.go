package deploy

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
)

type errMsg error

type logMsg struct {
	timestamp time.Time
	log       string
	stream    string
}

type spinnerModel struct {
	spinner      spinner.Model
	quitting     bool
	err          error
	start        time.Time
	lines        int
	displayedLogs map[string]bool // Track which logs have been displayed

	logs []logMsg
}

func initialSpinnerModel() spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return spinnerModel{
		spinner:       s,
		start:         time.Now(),
		lines:         0,
		displayedLogs: make(map[string]bool),
	}
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle Ctrl+C to allow graceful interruption
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}

	case tea.QuitMsg:
		m.quitting = true
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil

	case logMsg:
		// Create a unique key for the log entry
		logKey := fmt.Sprintf("%s_%s_%s", msg.timestamp.Format(time.RFC3339Nano), msg.stream, msg.log)
		
		// Only add if we haven't seen this log before
		if !m.displayedLogs[logKey] {
			m.logs = append(m.logs, msg)
			m.displayedLogs[logKey] = true
		}
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	if m.quitting {
		return ""
	}

	var output string

	// Sort logs by timestamp before displaying
	sort.Slice(m.logs, func(i, j int) bool {
		return m.logs[i].timestamp.Before(m.logs[j].timestamp)
	})

	// Display all logs first
	for _, log := range m.logs {
		if log.stream == "stdout" {
			output += color.New(color.Faint).Sprint(log.log) + "\n"
		} else {
			output += color.RedString(log.log) + "\n"
		}
	}
	
	// Add the spinner status at the bottom
	elapsed := time.Since(m.start).Round(time.Second)
	output += fmt.Sprintf("%s Deploying... (%s)", m.spinner.View(), elapsed)
	
	return output
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

