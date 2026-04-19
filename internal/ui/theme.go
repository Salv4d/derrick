package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// DebugMode enables verbose debug output and raw command logs.
var DebugMode bool

// Quiet suppresses decorative output (headers, sections, info, success,
// warning) so machine-readable formats like --json remain pure. Errors
// and FailFast still print since they are diagnostically essential.
var Quiet bool

var (
	colorSuccess = lipgloss.Color("46")
	colorError   = lipgloss.Color("196")
	colorInfo    = lipgloss.Color("39")
	colorWarning = lipgloss.Color("214")
	colorDebug   = lipgloss.Color("244")

	styleSuccess = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	styleError   = lipgloss.NewStyle().Foreground(colorError).Bold(true)
	styleInfo    = lipgloss.NewStyle().Foreground(colorInfo)
	styleWarning = lipgloss.NewStyle().Foreground(colorWarning).Italic(true)
	styleDebug   = lipgloss.NewStyle().Foreground(colorDebug).Italic(true)

	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2).
			MarginBottom(1)

	styleSection = lipgloss.NewStyle().Bold(true).MarginTop(1).MarginBottom(1)
	styleTask    = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Align(lipgloss.Left)
	styleSubTask = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

// PrintHeader prints the Derrick CLI header.
func PrintHeader() {
	if Quiet {
		return
	}
	fmt.Println(styleHeader.Render("DERRICK CLI"))
}

// Section prints a section header.
func Section(msg string) {
	if Quiet {
		return
	}
	fmt.Println(styleSection.Render("━━ " + msg))
}

// Sectionf prints a formatted section header.
func Sectionf(format string, args ...any) {
	Section(fmt.Sprintf(format, args...))
}

// Task prints a task indicator.
func Task(msg string) {
	if Quiet {
		return
	}
	fmt.Printf("  %s\n", styleTask.Render(msg+"..."))
}

// Taskf prints a formatted task indicator.
func Taskf(format string, args ...any) {
	Task(fmt.Sprintf(format, args...))
}

// SubTask prints a subtask indicator.
func SubTask(msg string) {
	if Quiet {
		return
	}
	fmt.Printf("    %s ", styleSubTask.Render(msg+"..."))
}

// SubTaskf prints a formatted subtask indicator.
func SubTaskf(format string, args ...any) {
	SubTask(fmt.Sprintf(format, args...))
}

// Info prints an informational message.
func Info(msg string) {
	if Quiet {
		return
	}
	fmt.Println(styleInfo.Render("ℹ  " + msg))
}

// Success prints a success message.
func Success(msg string) {
	if Quiet {
		return
	}
	fmt.Println(styleSuccess.Render("✓  " + msg))
}

// Warning prints a warning message.
func Warning(msg string) {
	if Quiet {
		return
	}
	fmt.Println(styleWarning.Render("⚠  " + msg))
}

// Error prints an error message.
func Error(msg string) { fmt.Println(styleError.Render("✖  " + msg)) }

// Debug prints a debug message if DebugMode is enabled.
func Debug(msg string) {
	if DebugMode {
		fmt.Println(styleDebug.Render("⚙ [DEBUG] " + msg))
	}
}

// Infof prints a formatted informational message.
func Infof(format string, args ...any) {
	if Quiet {
		return
	}
	fmt.Println(styleInfo.Render("ℹ  " + fmt.Sprintf(format, args...)))
}

// Successf prints a formatted success message.
func Successf(format string, args ...any) {
	if Quiet {
		return
	}
	fmt.Println(styleSuccess.Render("✓  " + fmt.Sprintf(format, args...)))
}

// Warningf prints a formatted warning message.
func Warningf(format string, args ...any) {
	if Quiet {
		return
	}
	fmt.Println(styleWarning.Render("⚠  " + fmt.Sprintf(format, args...)))
}

// Errorf prints a formatted error message.
func Errorf(format string, args ...any) {
	fmt.Println(styleError.Render("✖  " + fmt.Sprintf(format, args...)))
}

// Debugf prints a formatted debug message if DebugMode is enabled.
func Debugf(format string, args ...any) {
	if DebugMode {
		fmt.Println(styleDebug.Render("⚙ [DEBUG] " + fmt.Sprintf(format, args...)))
	}
}

// SprintSuccess returns a formatted success message string.
func SprintSuccess(format string, args ...any) string {
	return styleSuccess.Render("✓  " + fmt.Sprintf(format, args...))
}

// SprintError returns a formatted error message string.
func SprintError(format string, args ...any) string {
	return styleError.Render("✖  " + fmt.Sprintf(format, args...))
}

// SprintWarning returns a formatted warning message string.
func SprintWarning(format string, args ...any) string {
	return styleWarning.Render("⚠  " + fmt.Sprintf(format, args...))
}

// SprintInfo returns a formatted informational message string.
func SprintInfo(format string, args ...any) string {
	return styleInfo.Render("ℹ  " + fmt.Sprintf(format, args...))
}

// FailFast prints a critical error and exits.
func FailFast(err error) {
	fmt.Printf("\n%s\n", styleError.Render("✖ CRITICAL ERROR"))
	fmt.Println(styleError.Render(err.Error()))
	os.Exit(1)
}

// FailFastf prints a formatted critical error and exits.
func FailFastf(format string, args ...any) {
	fmt.Printf("\n%s\n", styleError.Render("✖ CRITICAL ERROR"))
	fmt.Println(styleError.Render(fmt.Sprintf(format, args...)))
	os.Exit(1)
}
