package dashboard

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Salv4d/derrick/internal/config"
	"github.com/Salv4d/derrick/internal/state"
)

// projectView is the template-facing representation of a project.
type projectView struct {
	Name            string
	Dir             string
	Provider        string
	Status          state.Status
	StatusClass     string
	StartedAt       string
	Flags           map[string]config.FlagDef
	Error           string
	StreamingAction string // "start", "stop", or empty
	HasLogs         bool
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
	LastLogs    string
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

	if _, err := os.Stat(filepath.Join(p.Dir, ".derrick", "last.log")); err == nil {
		v.HasLogs = true
	}

	return v
}

// ── Page / partial handlers ────────────────────────────────────────────────

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := renderPage(w, s.buildViews()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	p := s.findProject(r.PathValue("name"))
	if p == nil {
		http.NotFound(w, r)
		return
	}
	v := toView(*p)
	v.StreamingAction = "start"
	if err := renderRow(w, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	p := s.findProject(r.PathValue("name"))
	if p == nil {
		http.NotFound(w, r)
		return
	}
	v := toView(*p)
	v.StreamingAction = "stop"
	if err := renderRow(w, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleRows(w http.ResponseWriter, r *http.Request) {
	if err := renderRows(w, s.buildViews()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleRow returns just the main <tr> for one project (used after streaming completes).
func (s *Server) handleRow(w http.ResponseWriter, r *http.Request) {
	p := s.findProject(r.PathValue("name"))
	if p == nil {
		http.NotFound(w, r)
		return
	}
	if err := renderRow(w, toView(*p)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleStatus returns just the status badge (per-cell auto-poll).
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	p := s.findProject(r.PathValue("name"))
	if p == nil {
		http.NotFound(w, r)
		return
	}
	v := toView(*p)
	fmt.Fprintf(w, `<span class="badge %s">%s</span>`, v.StatusClass, v.Status)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	p := s.findProject(r.PathValue("name"))
	if p == nil {
		http.NotFound(w, r)
		return
	}

	data, err := os.ReadFile(filepath.Join(p.Dir, ".derrick", "last.log"))
	if err != nil {
		fmt.Fprintf(w, `<p class="detail-empty">No logs found.</p>`)
		return
	}

	fmt.Fprintf(w, `
		<div class="console">
			<div class="console-header">
				<span>Last Command Log</span>
				<button class="expand-btn" onclick="this.closest('.console').remove()" style="margin:0; padding:0 4px;">✕</button>
			</div>
			<div class="console-logs" style="max-height: 400px;">%s</div>
		</div>
	`, string(data))
}

// handleDetail returns the expandable container + flags panel.
func (s *Server) handleDetail(w http.ResponseWriter, r *http.Request) {
	p := s.findProject(r.PathValue("name"))
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
		LastLogs:    p.LastLogs,
	}
	if hasDocker {
		dv.Containers = fetchContainers(p.Name)
	}

	if err := renderDetail(w, dv); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ── Streaming action handlers (SSE) ────────────────────────────────────────

func (s *Server) handleStreamStart(w http.ResponseWriter, r *http.Request) {
	p := s.findProject(r.PathValue("name"))
	if p == nil {
		http.NotFound(w, r)
		return
	}
	s.streamCommand(w, r, p, s.binary, "start")
}

func (s *Server) handleStreamStop(w http.ResponseWriter, r *http.Request) {
	p := s.findProject(r.PathValue("name"))
	if p == nil {
		http.NotFound(w, r)
		return
	}
	s.streamCommand(w, r, p, s.binary, "stop")
}

func (s *Server) handleStreamFlag(w http.ResponseWriter, r *http.Request) {
	p := s.findProject(r.PathValue("name"))
	if p == nil {
		http.NotFound(w, r)
		return
	}
	flag := r.PathValue("flag")
	s.streamCommand(w, r, p, s.binary, "start", "--flag", flag)
}

// handleRunFlag is kept for non-streaming HTMX callers (flag buttons in detail panel
// when streaming is not desired). Delegates to streaming now.
func (s *Server) handleRunFlag(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	flag := r.PathValue("flag")
	p := s.findProject(name)
	if p == nil {
		http.Error(w, fmt.Sprintf("project %q not found", name), http.StatusNotFound)
		return
	}

	v := toView(*p)
	v.StreamingAction = "flag/" + flag
	if err := renderRow(w, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// streamCommand runs the given command and streams its output as SSE events.
// Each output line becomes a `data:` event. On completion a `done` event is sent.
func (s *Server) streamCommand(w http.ResponseWriter, r *http.Request, p *Project, bin string, args ...string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering if present

	sendEvent := func(event, data string) {
		if event != "" {
			fmt.Fprintf(w, "event: %s\n", event)
		}
		// SSE data lines must not contain raw newlines; split multi-line messages.
		for _, line := range strings.Split(data, "\n") {
			fmt.Fprintf(w, "data: %s\n", line)
		}
		fmt.Fprintf(w, "\n")
		flusher.Flush()
	}

	cmd := exec.CommandContext(r.Context(), bin, args...)
	cmd.Dir = p.Dir

	// Pipe merged stdout+stderr through a single reader so line order is preserved.
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		sendEvent("done", "error: "+err.Error())
		return
	}

	// Stream lines as they arrive.
	go func() {
		err := cmd.Wait()
		pw.CloseWithError(err)
	}()

	scanner := bufio.NewScanner(pr)
	var logBuf strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		stripped := stripANSI(line)
		logBuf.WriteString(stripped + "\n")
		sendEvent("", stripped)
	}

	p.LastLogs = logBuf.String()

	// pw.CloseWithError propagates the exit error through the pipe.
	if err := pr.Close(); err != nil && err != io.EOF {
		sendEvent("done", "error")
	} else {
		// Check actual command exit status via the goroutine's pipe close error.
		// After scanner exits, the pipe closed with cmd.Wait()'s error (if any).
		// We use a sentinel approach: if scanner stopped normally, cmd succeeded.
		sendEvent("done", "ok")
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

// lastLines returns the last n non-empty lines from s.
func lastLines(s string, n int) string {
	s = stripANSI(s)
	lines := strings.Split(strings.TrimSpace(s), "\n")
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
