package internalhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage"
)

type Server struct {
	logger  Logger
	app     Application
	host    string
	port    int
	httpSrv *http.Server
}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
}

type Application interface {
	CreateEvent(ctx context.Context, e storage.Event) error
	UpdateEvent(ctx context.Context, e storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEvent(ctx context.Context, id string) (storage.Event, error)
	ListEvents(ctx context.Context) ([]storage.Event, error)
	ListEventsDay(ctx context.Context, dayStart time.Time) ([]storage.Event, error)
	ListEventsWeek(ctx context.Context, weekStart time.Time) ([]storage.Event, error)
	ListEventsMonth(ctx context.Context, monthStart time.Time) ([]storage.Event, error)
}

func NewServer(logger Logger, app Application, host string, port int) *Server {
	s := &Server{
		logger: logger,
		app:    app,
		host:   host,
		port:   port,
	}

	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.createEventHandler(w, r)
		case http.MethodGet:
			s.listEventsHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/events/update", s.updateEventHandler)
	mux.HandleFunc("/api/events/delete", s.deleteEventHandler)
	mux.HandleFunc("/api/events/get", s.getEventHandler)
	mux.HandleFunc("/api/events/day", s.listEventsDayHandler)
	mux.HandleFunc("/api/events/week", s.listEventsWeekHandler)
	mux.HandleFunc("/api/events/month", s.listEventsMonthHandler)

	// Legacy endpoints for backward compatibility
	mux.HandleFunc("/hello", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("hello\n"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok\n"))
	})

	// wrap middleware
	handler := loggingMiddleware(mux, logger)

	s.httpSrv = &http.Server{
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return s
}

func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.logger.Info("http server listening on " + addr)

	go func() {
		if err := s.httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
			s.logger.Error("http serve error: " + err.Error())
		}
	}()

	<-ctx.Done()
	return s.Stop(context.Background())
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("shutting down http server")
	ctxShut, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.httpSrv.Shutdown(ctxShut); err != nil {
		s.logger.Error("graceful shutdown failed: " + err.Error())
		return err
	}
	return nil
}
