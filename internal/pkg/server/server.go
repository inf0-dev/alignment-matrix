package server

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"

	v1 "github.com/inf0-dev/alignment-matrix/api/v1"
	"github.com/inf0-dev/alignment-matrix/internal/pkg/engine"
	"github.com/inf0-dev/alignment-matrix/internal/pkg/parser"
	"github.com/inf0-dev/alignment-matrix/internal/pkg/renderer"
)

//go:embed templates/upload.html.tmpl
var uploadPage string

//go:embed templates/*
var _ embed.FS

// Config holds the server configuration.
type Config struct {
	Addr string
}

// Server serves an alignment matrix record as an interactive HTML page.
type Server struct {
	cfg    Config
	mu     sync.RWMutex
	record *v1.Record // nil when no document loaded
}

// New creates a new Server. Record may be nil to start without a document.
func New(cfg Config, record *v1.Record) *Server {
	return &Server{
		cfg:    cfg,
		record: record,
	}
}

// Run starts the HTTP server and blocks until the context is cancelled.
func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/upload", s.handleUpload)
	mux.HandleFunc("/", s.handlePage)

	srv := &http.Server{
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	ln, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.cfg.Addr, err)
	}

	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()

	log.Printf("Serving on http://%s", ln.Addr())

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (s *Server) handlePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	s.mu.RLock()
	rec := s.record
	s.mu.RUnlock()

	if rec == nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(uploadPage))
		return
	}

	html, err := renderer.HTML(rec)
	if err != nil {
		http.Error(w, "render failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	const maxSize = 2 << 20 // 2 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	if err := r.ParseMultipartForm(maxSize); err != nil {
		http.Error(w, "file too large (max 2 MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	rec, err := parseUpload(data, header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.record = rec
	s.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func parseUpload(data []byte, filename string) (*v1.Record, error) {
	docType, err := parser.DetectType(data, filename)
	if err != nil {
		return nil, err
	}

	switch docType {
	case "document":
		doc, err := parser.ParseDocument(data, filename)
		if err != nil {
			return nil, fmt.Errorf("failed to parse document: %w", err)
		}
		if err := doc.Validate(); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}
		return engine.Evaluate(doc, nil, nil)
	case "record":
		rec, err := parser.ParseRecord(data, filename)
		if err != nil {
			return nil, fmt.Errorf("failed to parse record: %w", err)
		}
		if err := rec.Validate(); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}
		return rec, nil
	default:
		return nil, fmt.Errorf("invalid type: %s", docType)
	}
}
