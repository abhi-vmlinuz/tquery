package pager

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"golang.org/x/term"
)

type model struct {
	viewport viewport.Model
	ready    bool
	content  string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "g":
			m.viewport.GotoTop()
			return m, nil
		case "G":
			m.viewport.GotoBottom()
			return m, nil
		}

	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height)
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing pager..."
	}
	return m.viewport.View()
}

// ShouldPage determines whether output should be directed through a pager.
func ShouldPage(lineCount int, noPager bool) bool {
	if noPager {
		return false
	}

	// Only page if stdout is an interactive terminal
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return false
	}

	// Check TERM environment
	termEnv := os.Getenv("TERM")
	if termEnv == "dumb" || termEnv == "" {
		return false
	}

	// Get terminal height
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || height <= 0 {
		return false
	}

	// Page if content exceeds available terminal height
	return lineCount > (height - 1)
}

// PageOutput pipes content to custom $PAGER or launches the built-in Bubble Tea pager.
func PageOutput(content string) error {
	// If user explicitly configured TQ_PAGER or PAGER, try system pager
	if customPager := os.Getenv("TQ_PAGER"); customPager != "" {
		return runSystemPager(customPager, content)
	}
	if envPager := os.Getenv("PAGER"); envPager != "" && envPager != "less" {
		return runSystemPager(envPager, content)
	}

	// Default: Built-in Bubble Tea pager with AltScreen + Mouse Motion
	return RunBuiltinPager(content)
}

func runSystemPager(pagerCmd string, content string) error {
	if pagerCmd == "cat" {
		_, err := fmt.Print(content)
		return err
	}

	parts := strings.Fields(pagerCmd)
	binary := parts[0]
	args := parts[1:]

	cmd := exec.Command(binary, args...)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if strings.Contains(binary, "less") {
		cmd.Env = append(os.Environ(), "LESSCHARSET=utf-8")
	}

	if err := cmd.Run(); err != nil {
		_, _ = fmt.Print(content)
	}
	return nil
}

// RunBuiltinPager runs the Bubble Tea viewport pager with alternate screen & mouse capture.
func RunBuiltinPager(content string) error {
	m := model{content: content}
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),       // Clean alternate screen (no scrollback leak!)
		tea.WithMouseCellMotion(), // Mouse wheel scrolls viewport, NOT terminal scrollbar!
	)
	_, err := p.Run()
	return err
}

// WriteOrPage writes content to writer or launches pager depending on conditions.
func WriteOrPage(w io.Writer, content string, noPager bool) error {
	lineCount := strings.Count(content, "\n")
	if ShouldPage(lineCount, noPager) {
		return PageOutput(content)
	}

	_, err := io.WriteString(w, content)
	return err
}
