package dashboard

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Salv4d/derrick/internal/config"
)

// Project is a derrick project the dashboard knows about.
type Project struct {
	Name     string
	Dir      string
	Config   *config.ProjectConfig
	LastLogs string
}

// Server is the dashboard HTTP server.
type Server struct {
	projects []Project
	binary   string
	mux      *http.ServeMux
}

// New builds a Server from a list of project directories. Each dir must contain
// a derrick.yaml. If dirs is empty the current working directory is used.
func New(dirs []string) (*Server, error) {
	if len(dirs) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("dashboard: could not determine working directory: %w", err)
		}
		dirs = []string{cwd}
	}

	projects := make([]Project, 0, len(dirs))
	for _, d := range dirs {
		abs, err := filepath.Abs(d)
		if err != nil {
			return nil, fmt.Errorf("dashboard: bad path %q: %w", d, err)
		}
		cfg, err := config.ParseConfig(filepath.Join(abs, "derrick.yaml"), "")
		if err != nil {
			return nil, fmt.Errorf("dashboard: %s: %w", d, err)
		}
		projects = append(projects, Project{Name: cfg.Name, Dir: abs, Config: cfg})
	}

	s := &Server{
		projects: projects,
		binary:   os.Args[0],
		mux:      http.NewServeMux(),
	}
	s.registerRoutes()
	return s, nil
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /", s.handleIndex)
	s.mux.HandleFunc("GET /rows", s.handleRows)
	s.mux.HandleFunc("GET /projects/{name}/row", s.handleRow)
	s.mux.HandleFunc("GET /projects/{name}/status", s.handleStatus)
	s.mux.HandleFunc("GET /projects/{name}/detail", s.handleDetail)
	s.mux.HandleFunc("GET /projects/{name}/logs", s.handleLogs)
	s.mux.HandleFunc("GET /projects/{name}/stream/start", s.handleStreamStart)
	s.mux.HandleFunc("GET /projects/{name}/stream/stop", s.handleStreamStop)
	s.mux.HandleFunc("GET /projects/{name}/stream/flag/{flag}", s.handleStreamFlag)

	s.mux.HandleFunc("POST /projects/{name}/start", s.handleStart)
	s.mux.HandleFunc("POST /projects/{name}/stop", s.handleStop)
	// Keep POST routes for direct HTMX callers (flags from detail panel).
	s.mux.HandleFunc("POST /projects/{name}/flag/{flag}", s.handleRunFlag)
}

// Serve starts the HTTP server on the given port and blocks until ctx is cancelled.
func (s *Server) Serve(ctx context.Context, port int) error {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("dashboard: cannot listen on %s: %w", addr, err)
	}

	srv := &http.Server{Handler: s.mux}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	fmt.Printf("\n  Derrick Dashboard → http://localhost:%d\n\n", port)
	err = srv.Serve(ln)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// findProject returns the Project with the given name, or nil.
func (s *Server) findProject(name string) *Project {
	for i := range s.projects {
		if s.projects[i].Name == name {
			return &s.projects[i]
		}
	}
	return nil
}
