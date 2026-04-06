package ui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

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

// DashboardModel is the Bubble Tea model for the interactive TUI dashboard.
type DashboardModel struct {
	activeTab  int
	tabs       []string
	width      int
	height     int
	containers []Container
	logs       []string
	logScanner *bufio.Scanner
}

// Container represents a Docker container for display in the dashboard.
type Container struct {
	Names  string `json:"Names"`
	Status string `json:"Status"`
	State  string `json:"State"`
}

type dockerStatusMsg []Container

// NewDashboardModel creates a new dashboard model with default values.
func NewDashboardModel() DashboardModel {
	return DashboardModel{
		activeTab: 0,
		tabs:      []string{"Containers", "Logs", "Config"},
		logs:      []string{"System initialized. Awaiting logs..."},
	}
}

// Init initializes the dashboard model and starts background commands.
func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(fetchDocker, startLogStream)
}

type logStreamInitMsg struct {
	scanner *bufio.Scanner
}
type logLineMsg string

func startLogStream() tea.Msg {
	cmd := exec.Command("docker", "compose", "logs", "-f", "--tail", "20")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return logLineMsg("Error: could not pipe logs. Are you in a compose directory?")
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return logLineMsg("Error: failed to start log sub-process.")
	}

	return logStreamInitMsg{
		scanner: bufio.NewScanner(stdout),
	}
}

func waitForNextLog(scanner *bufio.Scanner) tea.Cmd {
	return func() tea.Msg {
		if scanner.Scan() {
			return logLineMsg(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return logLineMsg(fmt.Sprintf("Stream scan err: %v", err))
		}
		return logLineMsg("[Stream Disconnected]")
	}
}

func fetchDocker() tea.Msg {
	cmd := exec.Command("docker", "ps", "--format", "{{json .}}")
	out, err := cmd.Output()
	if err != nil {
		return dockerStatusMsg{}
	}

	var containers []Container
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var c Container
		if err := json.Unmarshal([]byte(line), &c); err == nil {
			containers = append(containers, c)
		}
	}
	return dockerStatusMsg(containers)
}

func tickDockerStatus() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return fetchDocker()
	})
}

// Update handles incoming messages and updates the model state.
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
	case dockerStatusMsg:
		m.containers = msg
		return m, tickDockerStatus()
	case logStreamInitMsg:
		m.logScanner = msg.scanner
		return m, waitForNextLog(m.logScanner)
	case logLineMsg:
		m.logs = append(m.logs, string(msg))
		if len(m.logs) > m.height-10 {
			m.logs = m.logs[1:]
		}
		if m.logScanner != nil {
			return m, waitForNextLog(m.logScanner)
		}
		return m, nil
	}
	return m, nil
}

// View renders the dashboard UI.
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

		content := ""
	switch m.activeTab {
	case 0:
		if len(m.containers) == 0 {
			content = "Container Status: [No running containers found or scanning...]\n"
		} else {
			sb := strings.Builder{}
			sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Running Containers:"))
			sb.WriteString("\n\n")
			for _, c := range m.containers {
				statusColor := "42"
				if c.State != "running" {
					statusColor = "196"
				}
				statusTag := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render(fmt.Sprintf("[%s]", c.State))
				sb.WriteString(fmt.Sprintf("%s %s - %s\n", statusTag, lipgloss.NewStyle().Bold(true).Render(c.Names), c.Status))
			}
			content = sb.String()
		}
		content += "\n(Press 'q' or 'ctrl+c' to exit, 'tab' to change menus)"
	case 1:
		sb := strings.Builder{}
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Real-time Compose Logs:"))
		sb.WriteString("\n\n")
		for _, l := range m.logs {
			sb.WriteString(l + "\n")
		}
		content = sb.String()
	case 2:
		content = "Global Configuration:\n[Registry Sync Pending...]"
	}

	doc.WriteString(content)

	return baseStyle.Render(doc.String())
}
