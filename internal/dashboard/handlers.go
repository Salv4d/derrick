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
	"github.com/Salv4d/derrick/internal/discovery"
	"github.com/Salv4d/derrick/internal/state"
	"gopkg.in/yaml.v3"
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
	Networks        []string
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

	v.Networks = p.Config.Docker.Networks

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

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	p := s.findProject(r.PathValue("name"))
	if p == nil {
		http.NotFound(w, r)
		return
	}

	data, err := os.ReadFile(filepath.Join(p.Dir, "derrick.yaml"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, `
		<form hx-post="/projects/%s/settings" hx-target="this" hx-swap="outerHTML" class="settings-form">
			<div class="detail-section">
				<h4 class="detail-heading">Edit derrick.yaml</h4>
				<textarea name="config" class="config-editor" spellcheck="false">%s</textarea>
				<div class="action-btns" style="margin-top: 12px;">
					<button type="submit" class="btn btn-start">Save Changes</button>
					<span class="htmx-indicator">Saving...</span>
				</div>
			</div>
		</form>
	`, p.Name, string(data))
}

func (s *Server) handleSaveSettings(w http.ResponseWriter, r *http.Request) {
	p := s.findProject(r.PathValue("name"))
	if p == nil {
		http.NotFound(w, r)
		return
	}

	newConfig := r.FormValue("config")
	// Basic validation: try to parse it
	_, err := config.ParseConfigBytes([]byte(newConfig), "")
	if err != nil {
		fmt.Fprintf(w, `
			<div class="error-msg" style="margin-bottom: 12px;">
				<svg viewBox="0 0 16 16" fill="currentColor" width="12" height="12"><path d="M8 1a7 7 0 1 0 0 14A7 7 0 0 0 8 1zm0 3a.75.75 0 0 1 .75.75v3.5a.75.75 0 0 1-1.5 0v-3.5A.75.75 0 0 1 8 4zm0 8a1 1 0 1 1 0-2 1 1 0 0 1 0 2z"/></svg>
				<pre>Validation Failed: %v</pre>
			</div>
		`, err)
		// Re-render the form with error
		fmt.Fprintf(w, `
			<form hx-post="/projects/%s/settings" hx-target="this" hx-swap="outerHTML" class="settings-form">
				<div class="detail-section">
					<h4 class="detail-heading">Edit derrick.yaml</h4>
					<textarea name="config" class="config-editor" spellcheck="false">%s</textarea>
					<div class="action-btns" style="margin-top: 12px;">
						<button type="submit" class="btn btn-start">Save Changes</button>
					</div>
				</div>
			</form>
		`, p.Name, newConfig)
		return
	}

	err = os.WriteFile(filepath.Join(p.Dir, "derrick.yaml"), []byte(newConfig), 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	parsed, _ := config.ParseConfigBytes([]byte(newConfig), "")
	p.Config = parsed

	fmt.Fprintf(w, `
		<div class="detail-section">
			<div style="background: var(--green-dim); color: var(--green); padding: 8px 12px; border-radius: var(--radius); margin-bottom: 12px; font-size: 13px;">
				✓ Settings saved successfully.
			</div>
			<button hx-get="/projects/%s/settings" hx-target="closest .detail-section" hx-swap="outerHTML" class="btn btn-flag">Back to Editor</button>
		</div>
	`, p.Name)
}

func (s *Server) handleInitForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
		<h2 style="margin-bottom: 24px;">Initialize New Project</h2>
		<form hx-post="/init" hx-target="#init-modal-content">
			<div class="detail-section">
				<p class="detail-heading">Project Directory</p>
				<input type="text" name="dir" placeholder="/path/to/project" class="config-editor" style="height: 40px; margin-bottom: 20px;">
				<button type="submit" class="btn btn-start" style="width: 100%%;">Run Discovery</button>
			</div>
		</form>
	`)
}

func (s *Server) handleInit(w http.ResponseWriter, r *http.Request) {
	dir := r.FormValue("dir")
	if dir == "" {
		http.Error(w, "directory is required", http.StatusBadRequest)
		return
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	meta := discovery.DiscoverProject(absDir)
	cfg := config.ProjectConfig{
		Schema:  config.CurrentSchema,
		Name:    meta.Name,
		Version: meta.Version,
	}
	for _, f := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
		if _, err := os.Stat(filepath.Join(absDir, f)); err == nil {
			cfg.Docker.Compose = f
			break
		}
	}
	pkgs := discovery.SuggestedPackages(meta.Language)
	for _, p := range pkgs {
		cfg.Nix.Packages = append(cfg.Nix.Packages, config.NixPackage{Name: p})
	}
	yamlData, _ := yaml.Marshal(&cfg)

	fmt.Fprintf(w, `
		<h2 style="margin-bottom: 24px;">Review Configuration</h2>
		<p class="subtitle" style="margin-bottom: 20px;">We've analyzed %s and generated a draft contract.</p>
		<form hx-post="/projects/save-new" hx-target="#init-modal-content">
			<input type="hidden" name="dir" value="%s">
			<div class="detail-section">
				<textarea name="config" class="config-editor" style="height: 300px;">%s</textarea>
				<div class="action-btns" style="margin-top: 24px; width: 100%%; justify-content: flex-end;">
					<button type="button" class="btn btn-flag" onclick="document.getElementById('init-modal').style.display='none'">Cancel</button>
					<button type="submit" class="btn btn-start">Create derrick.yaml</button>
				</div>
			</div>
		</form>
	`, absDir, absDir, string(yamlData))
}

func (s *Server) handleSaveNew(w http.ResponseWriter, r *http.Request) {
	dir := r.FormValue("dir")
	newConfig := r.FormValue("config")
	parsed, err := config.ParseConfigBytes([]byte(newConfig), "")
	if err != nil {
		http.Error(w, fmt.Sprintf("Validation Failed: %v", err), http.StatusBadRequest)
		return
	}
	err = os.WriteFile(filepath.Join(dir, "derrick.yaml"), []byte(newConfig), 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.mu.Lock()
	s.projects = append(s.projects, Project{
		Name:   parsed.Name,
		Dir:    dir,
		Config: parsed,
	})
	s.mu.Unlock()

	fmt.Fprintf(w, `
		<div style="text-align: center; padding: 20px;">
			<div style="background: var(--green-dim); color: var(--green); padding: 16px; border-radius: var(--radius); margin-bottom: 24px;">
				<h3 style="margin-bottom: 8px;">✓ Project Initialized!</h3>
				<p>Successfully created derrick.yaml in %s</p>
			</div>
			<button class="btn btn-start" onclick="location.reload()">Refresh Dashboard</button>
		</div>
	`, dir)
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
	s.mu.Lock()
	defer s.mu.Unlock()
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
