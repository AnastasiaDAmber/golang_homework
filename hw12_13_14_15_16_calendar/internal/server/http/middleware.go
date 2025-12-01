package internalhttp

import (
	"fmt"
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

		clientIP := ipFromRequest(r)
		ua := r.UserAgent()
		method := r.Method
		path := r.URL.RequestURI()
		proto := r.Proto
		status := lrw.status
		if status == 0 {
			status = 200
		}
		// Формат: IP [дата] метод путь версия код размер "user-agent"
		// Пример: 66.249.65.3 [25/Feb/2020:19:11:24 +0600] GET /hello?q=1 HTTP/1.1 200 30 "Mozilla/5.0"
		msg := clientIP + " [" + start.Format("02/Jan/2006:15:04:05 -0700") + "] " +
			method + " " + path + " " + proto + " " +
			fmt.Sprintf("%d", status) + " " +
			fmt.Sprintf("%d", lrw.size) + " \"" + ua + "\""

		logger.Info(msg)
	})
}

