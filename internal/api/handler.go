package api

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"strconv"

	"github.com/Mineru98/disk-viz-viewer/internal/disk"
)

// Server represents the HTTP server
type Server struct {
	staticFS embed.FS
}

// NewServer creates a new API server
func NewServer(staticFS embed.FS) *Server {
	return &Server{
		staticFS: staticFS,
	}
}

// SetupRoutes configures the HTTP routes
func (s *Server) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/analyze", s.handleAnalyze)

	// Static files
	staticContent, err := fs.Sub(s.staticFS, "web/static")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", http.FileServer(http.FS(staticContent)))

	return mux
}

// handleAnalyze handles the disk usage analysis API
func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	depthStr := r.URL.Query().Get("depth")
	depth := 1
	if depthStr != "" {
		if d, err := strconv.Atoi(depthStr); err == nil && d > 0 && d <= 5 {
			depth = d
		}
	}

	result, err := disk.AnalyzeDiskUsage(path, depth)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
