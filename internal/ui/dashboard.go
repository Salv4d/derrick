package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	baseStyle = lipgloss.NewStyle().
			Padding(1, 2)
	
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("205"))

	inactiveTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

type DashboardModel struct {
	activeTab int
	tabs      []string
	width     int
	height    int
}

func NewDashboardModel() DashboardModel {
	return DashboardModel{
		activeTab: 0,
		tabs:      []string{"Containers", "Logs", "Config"},
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return nil
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab", "right", "l":
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
		case "shift+tab", "left", "h":
			m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m DashboardModel) View() string {
	doc := strings.Builder{}

	var renderedTabs []string
	for i, tab := range m.tabs {
		if i == m.activeTab {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(tab))
		} else {
			renderedTabs = append(renderedTabs, inactiveTabStyle.Render(tab))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n\n")

	// Render Content
	content := ""
	switch m.activeTab {
	case 0:
		content = "Container Status: [Async Health Check Pending...]\n\n(Press 'q' or 'ctrl+c' to exit, 'tab' to change menus)"
	case 1:
		content = "Log Stream:\n[Multiplexer Async Data Pending...]"
	case 2:
		content = "Global Configuration:\n[Registry Sync Pending...]"
	}

	doc.WriteString(content)

	return baseStyle.Render(doc.String())
}
