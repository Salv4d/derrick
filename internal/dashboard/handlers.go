package dashboard

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/state"
)

// projectView is the template-facing representation of a project.
type projectView struct {
	Name        string
	Dir         string
	Provider    string
	Status      state.Status
	StatusClass string
	StartedAt   string
	Flags       map[string]config.FlagDef
	Error       string
}

// containerInfo holds display data for one running container.
type containerInfo struct {
	Name   string
	Image  string
	Status string
	IsUp   bool
}

// detailView is passed to the detail panel template.
type detailView struct {
	ProjectName string
	Containers  []containerInfo
	HasDocker   bool
	Flags       map[string]config.FlagDef
}

func toView(p Project) projectView {
	s, _ := state.Load(p.Dir)

	v := projectView{
		Name:     p.Name,
		Dir:      p.Dir,
		Provider: s.Provider,
		Status:   s.Status,
		Flags:    p.Config.Flags,
	}

	switch s.Status {
	case state.StatusRunning:
		v.StatusClass = "badge-running"
		if !s.StartedAt.IsZero() {
			v.StartedAt = humanDuration(time.Since(s.StartedAt))
		}
	case state.StatusStopped:
		v.StatusClass = "badge-stopped"
	default:
		v.StatusClass = "badge-unknown"
	}

	if v.Provider == "" {
		v.Provider = "—"
	}

	return v
}

// ── Page handlers ──────────────────────────────────────────────────────────

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	views := s.buildViews()
	if err := renderPage(w, views); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleRows(w http.ResponseWriter, r *http.Request) {
	views := s.buildViews()
	if err := renderRows(w, views); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleStatus returns just the status badge for a project (used by per-cell polling).
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	p := s.findProject(name)
	if p == nil {
		http.NotFound(w, r)
		return
	}
	v := toView(*p)
	fmt.Fprintf(w, `<span class="badge %s">%s</span>`, v.StatusClass, v.Status)
}

// handleDetail returns the expandable detail panel: containers + runnable flags.
func (s *Server) handleDetail(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	p := s.findProject(name)
	if p == nil {
		http.NotFound(w, r)
		return
	}

	provider := p.Config.ActiveProvider()
	hasDocker := provider == "docker" || provider == "hybrid"

	dv := detailView{
		ProjectName: p.Name,
		HasDocker:   hasDocker,
		Flags:       p.Config.Flags,
	}

	if hasDocker {
		dv.Containers = fetchContainers(p.Name)
	}

	if err := renderDetail(w, dv); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ── Action handlers ────────────────────────────────────────────────────────

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	s.runAction(w, r, "start")
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	s.runAction(w, r, "stop")
}

// handleRunFlag runs `derrick start --flag <flag>` for the given project.
func (s *Server) handleRunFlag(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	flag := r.PathValue("flag")
	p := s.findProject(name)
	if p == nil {
		http.Error(w, fmt.Sprintf("project %q not found", name), http.StatusNotFound)
		return
	}

	cmd := exec.Command(s.binary, "start", "--flag", flag)
	cmd.Dir = p.Dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()

	v := toView(*p)
	if err != nil {
		v.Error = lastLines(out.String(), 3)
	}
	if err := renderRow(w, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) runAction(w http.ResponseWriter, r *http.Request, action string) {
	name := r.PathValue("name")
	p := s.findProject(name)
	if p == nil {
		http.Error(w, fmt.Sprintf("project %q not found", name), http.StatusNotFound)
		return
	}

	cmd := exec.Command(s.binary, action)
	cmd.Dir = p.Dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()

	v := toView(*p)
	if err != nil {
		v.Error = lastLines(out.String(), 3)
	}
	if err := renderRow(w, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ── Helpers ────────────────────────────────────────────────────────────────

func (s *Server) buildViews() []projectView {
	views := make([]projectView, len(s.projects))
	for i, p := range s.projects {
		views[i] = toView(p)
	}
	return views
}

// fetchContainers queries docker ps for containers belonging to this compose project.
func fetchContainers(projectName string) []containerInfo {
	out, err := exec.Command(
		"docker", "ps",
		"--filter", "label=com.docker.compose.project="+projectName,
		"--format", "{{.Names}}|{{.Image}}|{{.Status}}",
	).Output()
	if err != nil || len(out) == 0 {
		return nil
	}

	var containers []containerInfo
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}
		containers = append(containers, containerInfo{
			Name:   parts[0],
			Image:  parts[1],
			Status: parts[2],
			IsUp:   strings.Contains(parts[2], "Up"),
		})
	}
	return containers
}

// lastLines returns the last n non-empty lines from s, stripped of ANSI.
func lastLines(s string, n int) string {
	// strip common ANSI escape sequences
	var cleaned strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++ // skip 'm'
			continue
		}
		cleaned.WriteByte(s[i])
		i++
	}

	lines := strings.Split(strings.TrimSpace(cleaned.String()), "\n")
	var nonempty []string
	for _, l := range lines {
		if t := strings.TrimSpace(l); t != "" {
			nonempty = append(nonempty, t)
		}
	}
	if len(nonempty) > n {
		nonempty = nonempty[len(nonempty)-n:]
	}
	return strings.Join(nonempty, "\n")
}

func humanDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}
