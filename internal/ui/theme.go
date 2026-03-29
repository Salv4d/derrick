package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	colorSuccess = lipgloss.Color("46")
	colorError = lipgloss.Color("196")
	colorInfo = lipgloss.Color("39")
	colorWarning = lipgloss.Color("214")
	
	styleSuccess = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	styleError = lipgloss.NewStyle().Foreground(colorError).Bold(true)
	styleInfo = lipgloss.NewStyle().Foreground(colorInfo)
	styleWarning = lipgloss.NewStyle().Foreground(colorWarning).Italic(true)

	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			PaddingTop(0).
			PaddingBottom(0).
			PaddingLeft(2).
			PaddingRight(2).
			MarginBottom(1)
)

func PrintHeader() {
	fmt.Println(styleHeader.Render("DERRICK CLI"))
}

func Info(msg string) {
	fmt.Println(styleInfo.Render("ℹ  " + msg))
}

func Success(msg string) {
	fmt.Println(styleSuccess.Render("✓  " + msg))
}

func Warning(msg string) {
	fmt.Println(styleWarning.Render("⚠  " + msg))
}

func FailFast(err error) {
	fmt.Printf("\n%s\n", styleError.Render("✖ CRITICAL ERROR"))
	fmt.Println(styleError.Render(err.Error()))
	os.Exit(1)
}

func SuccessInline(msg string) string {
	return styleSuccess.Render("✓ " + msg)
}

func ErrorInline(msg string) string {
	return styleError.Render("✖ " + msg)
}

func WarningInline(msg string) string {
	return styleWarning.Render("⚠ " + msg)
}