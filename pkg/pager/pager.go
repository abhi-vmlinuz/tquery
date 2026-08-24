package pager

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/mattn/go-isatty"
	"golang.org/x/term"
)

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

	// Only page if content exceeds available terminal height (leaving 1 line for prompt)
	return lineCount > (height - 1)
}

// PageOutput pipes the given content to a terminal pager (less -R or $PAGER).
// Falls back to direct stdout printing if pager is unavailable.
func PageOutput(content string) error {
	pagerCmd := getPagerCommand()
	if pagerCmd == "" || pagerCmd == "cat" {
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

	// Ensure LESS flags and UTF-8 support are configured if using less
	if strings.Contains(binary, "less") {
		cmd.Env = append(os.Environ(), "LESSCHARSET=utf-8")
	}

	if err := cmd.Run(); err != nil {
		// Fall back to direct stdout on any pager execution error
		_, _ = fmt.Print(content)
		return nil
	}

	return nil
}

func getPagerCommand() string {
	if p := os.Getenv("TQ_PAGER"); p != "" {
		return p
	}
	if p := os.Getenv("PAGER"); p != "" {
		// If user set PAGER to 'less', ensure -R flag is applied for ANSI colors and alternate screen window
		if p == "less" {
			return "less -R"
		}
		return p
	}

	// Default pager: less -R (opens alternate screen window, closes on q)
	if _, err := exec.LookPath("less"); err == nil {
		return "less -R"
	}

	return ""
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
