package internalhttp

import (
	"net"
	"net/http"
	"strings"
	"time"
)

func ipFromRequest(r *http.Request) string {
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		// может содержать несколько ip
		parts := strings.Split(x, ",")
		return strings.TrimSpace(parts[0])
	}
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}
	return r.RemoteAddr
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (lrw *loggingResponseWriter) WriteHeader(status int) {
	lrw.status = status
	lrw.ResponseWriter.WriteHeader(status)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.status == 0 {
		lrw.status = 200
	}
	n, err := lrw.ResponseWriter.Write(b)
	lrw.size += n
	return n, err
}

func loggingMiddleware(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(lrw, r)
		latency := time.Since(start)

		clientIP := ipFromRequest(r)
		ua := r.UserAgent()
		method := r.Method
		path := r.URL.RequestURI()
		proto := r.Proto
		status := lrw.status
		if status == 0 {
			status = 200
		}
		msg := clientIP + " [" + start.Format("02/Jan/2006:15:04:05 -0700") + "] " +
			method + " " + path + " " + proto + " " +
			http.StatusText(status) + " " +
			"(" + stringStatus(status) + ") " +
			latency.String() + " \"" + ua + "\""

		logger.Info(msg)
	})
}

func stringStatus(status int) string {
	return http.StatusText(status)
}
