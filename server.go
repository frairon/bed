package bed

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/warthog618/gpiod"
)

//go:embed templates/*
var templates embed.FS

type templateData struct {
	Entries []*Entry
}

type Server struct {
	templates *template.Template
	wakeup    *WakeUp
}

func NewServer(wakeup *WakeUp) *Server {
	return &Server{
		templates: template.Must(template.New("").ParseFS(templates, "templates/*.html")),
		wakeup:    wakeup,
	}
}

func (s *Server) Run(ctx context.Context) error {
	r := mux.NewRouter()
	// Add your routes as needed
	srv := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: r,
	}

	r.HandleFunc("/", s.renderOverview)
	r.HandleFunc("/wake", s.wakeNow)
	errors := make(chan error, 1)
	go func() {
		defer close(errors)
		if err := srv.ListenAndServe(); err != nil {
			errors <- fmt.Errorf("error running server: %v", err)
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errors:
		if err != nil {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func (s *Server) renderOverview(w http.ResponseWriter, r *http.Request) {
	data := templateData{
		Entries: s.wakeup.Entries(),
	}

	s.templates.ExecuteTemplate(w, "index.html", data)
}

func (s *Server) wakeNow(w http.ResponseWriter, r *http.Request) {
	s.wakeup.HandlePush(gpiod.LineEvent{})
	w.Write([]byte("ok"))
}
