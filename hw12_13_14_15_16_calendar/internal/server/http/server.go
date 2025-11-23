package internalhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
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

type Application interface{}

func NewServer(logger Logger, app Application, host string, port int) *Server {
	mux := http.NewServeMux()

	// hello handler
	mux.HandleFunc("/hello", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("hello\n"))
	})

	// root
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok\n"))
	})

	// wrap middleware
	handler := loggingMiddleware(mux, logger)

	s := &http.Server{
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &Server{
		logger:  logger,
		app:     app,
		host:    host,
		port:    port,
		httpSrv: s,
	}
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
