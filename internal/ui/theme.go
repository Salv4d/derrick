package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var DebugMode bool

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

func PrintHeader() {
	fmt.Println(styleHeader.Render("DERRICK CLI"))
}

func Section(msg string) {
	fmt.Println(styleSection.Render("━━ " + msg))
}

func Sectionf(format string, args ...any) {
	Section(fmt.Sprintf(format, args...))
}

func Task(msg string) {
	fmt.Printf("  %s\n", styleTask.Render(msg+"..."))
}

func Taskf(format string, args ...any) {
	Task(fmt.Sprintf(format, args...))
}

func SubTask(msg string) {
	fmt.Printf("    %s ", styleSubTask.Render(msg+"..."))
}

func SubTaskf(format string, args ...any) {
	SubTask(fmt.Sprintf(format, args...))
}

func Info(msg string)    { fmt.Println(styleInfo.Render("ℹ  " + msg)) }
func Success(msg string) { fmt.Println(styleSuccess.Render("✓  " + msg)) }
func Warning(msg string) { fmt.Println(styleWarning.Render("⚠  " + msg)) }
func Error(msg string)   { fmt.Println(styleError.Render("✖  " + msg)) }

func Debug(msg string) {
	if DebugMode {
		fmt.Println(styleDebug.Render("⚙ [DEBUG] " + msg))
	}
}

func Infof(format string, args ...any) {
	fmt.Println(styleInfo.Render("ℹ  " + fmt.Sprintf(format, args...)))
}

func Successf(format string, args ...any) {
	fmt.Println(styleSuccess.Render("✓  " + fmt.Sprintf(format, args...)))
}

func Warningf(format string, args ...any) {
	fmt.Println(styleWarning.Render("⚠  " + fmt.Sprintf(format, args...)))
}

func Errorf(format string, args ...any) {
	fmt.Println(styleError.Render("✖  " + fmt.Sprintf(format, args...)))
}

func Debugf(format string, args ...any) {
	if DebugMode {
		fmt.Println(styleDebug.Render("⚙ [DEBUG] " + fmt.Sprintf(format, args...)))
	}
}

func SprintSuccess(format string, args ...any) string {
	return styleSuccess.Render("✓  " + fmt.Sprintf(format, args...))
}

func SprintError(format string, args ...any) string {
	return styleError.Render("✖  " + fmt.Sprintf(format, args...))
}

func SprintWarning(format string, args ...any) string {
	return styleWarning.Render("⚠  " + fmt.Sprintf(format, args...))
}

func SprintInfo(format string, args ...any) string {
	return styleInfo.Render("ℹ  " + fmt.Sprintf(format, args...))
}

func FailFast(err error) {
	fmt.Printf("\n%s\n", styleError.Render("✖ CRITICAL ERROR"))
	fmt.Println(styleError.Render(err.Error()))
	os.Exit(1)
}

func FailFastf(format string, args ...any) {
	fmt.Printf("\n%s\n", styleError.Render("✖ CRITICAL ERROR"))
	fmt.Println(styleError.Render(fmt.Sprintf(format, args...)))
	os.Exit(1)
}
