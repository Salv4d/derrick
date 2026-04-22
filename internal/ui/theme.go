package ui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DebugMode enables verbose debug output and raw command logs.
var DebugMode bool

// Quiet suppresses decorative output (headers, sections, info, success,
// warning) so machine-readable formats like --json remain pure. Errors
// and FailFast still print since they are diagnostically essential.
var Quiet bool

// LogWriter is where plain-text (ANSI-stripped) logs are written.
var LogWriter io.Writer

// SetLogFile opens (or creates) .derrick/last.log in the given project
// directory and sets it as the LogWriter.
func SetLogFile(projectDir string) error {
	dir := filepath.Join(projectDir, ".derrick")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, "last.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	LogWriter = f
	return nil
}

func logmsg(msg string) {
	if LogWriter != nil {
		fmt.Fprintln(LogWriter, stripANSI(msg))
	}
}

// stripANSI removes ANSI escape codes from s.
func stripANSI(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

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
	logmsg("DERRICK CLI")
	if Quiet {
		return
	}
	fmt.Println(styleHeader.Render("DERRICK CLI"))
}

// Section prints a section header.
func Section(msg string) {
	logmsg("━━ " + msg)
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
	logmsg("  " + msg + "...")
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
	logmsg("    " + msg + "...")
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
	logmsg("ℹ  " + msg)
	if Quiet {
		return
	}
	fmt.Println(styleInfo.Render("ℹ  " + msg))
}

// Success prints a success message.
func Success(msg string) {
	logmsg("✓  " + msg)
	if Quiet {
		return
	}
	fmt.Println(styleSuccess.Render("✓  " + msg))
}

// Warning prints a warning message.
func Warning(msg string) {
	logmsg("⚠  " + msg)
	if Quiet {
		return
	}
	fmt.Println(styleWarning.Render("⚠  " + msg))
}

// Error prints an error message.
func Error(msg string) {
	logmsg("✖  " + msg)
	fmt.Println(styleError.Render("✖  " + msg))
}

// Debug prints a debug message if DebugMode is enabled.
func Debug(msg string) {
	if DebugMode {
		logmsg("⚙ [DEBUG] " + msg)
		fmt.Println(styleDebug.Render("⚙ [DEBUG] " + msg))
	}
}

// Infof prints a formatted informational message.
func Infof(format string, args ...any) {
	Info(fmt.Sprintf(format, args...))
}

// Successf prints a formatted success message.
func Successf(format string, args ...any) {
	Success(fmt.Sprintf(format, args...))
}

// Warningf prints a formatted warning message.
func Warningf(format string, args ...any) {
	Warning(fmt.Sprintf(format, args...))
}

// Errorf prints a formatted error message.
func Errorf(format string, args ...any) {
	Error(fmt.Sprintf(format, args...))
}

// Debugf prints a formatted debug message if DebugMode is enabled.
func Debugf(format string, args ...any) {
	Debug(fmt.Sprintf(format, args...))
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
	logmsg("✖ CRITICAL ERROR: " + err.Error())
	fmt.Printf("\n%s\n", styleError.Render("✖ CRITICAL ERROR"))
	fmt.Println(styleError.Render(err.Error()))
	os.Exit(1)
}

// FailFastf prints a formatted critical error and exits.
func FailFastf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	logmsg("✖ CRITICAL ERROR: " + msg)
	fmt.Printf("\n%s\n", styleError.Render("✖ CRITICAL ERROR"))
	fmt.Println(styleError.Render(msg))
	os.Exit(1)
}
