package dashboard

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"time"

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
	FlagsUsed   []string
}

func toView(p Project) projectView {
	s, _ := state.Load(p.Dir)

	v := projectView{
		Name:      p.Name,
		Dir:       p.Dir,
		Provider:  s.Provider,
		Status:    s.Status,
		FlagsUsed: s.FlagsUsed,
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

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	s.runAction(w, r, "start")
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	s.runAction(w, r, "stop")
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
	_ = cmd.Run() // errors surface in the updated state row

	v := toView(*p)
	if err := renderRow(w, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) buildViews() []projectView {
	views := make([]projectView, len(s.projects))
	for i, p := range s.projects {
		views[i] = toView(p)
	}
	return views
}

func humanDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", h, m)
}
